// file: cmd/audio-dedup/main.go
// package: main
//
// 命令行入口，解析参数，扫描源目录，计算指纹并去重，最后把保留的文件复制到目标目录。
// 运行示例（在项目根目录下）：
//
//	go run ./cmd/audio-dedup -src /path/to/src -dst /path/to/dst -workers 4 -threshold 8
package main

import (
	"deduplicateMusic/internal/copyutil"
	"deduplicateMusic/internal/dedup"
	"deduplicateMusic/internal/fingerprint"
	"deduplicateMusic/internal/report"
	"deduplicateMusic/internal/scanner"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// reportItems 用于存放每个文件的处理记录
var reportItems []report.ReportItem

func main() {
	// CLI 参数
	srcDir := flag.String("src", "", "源目录，包含待去重的音频文件")
	dstDir := flag.String("dst", "", "目标输出目录，保留的文件会被复制到此处")
	workers := flag.Int("workers", runtime.NumCPU(), "并发工作数量（默认：CPU 核数）")
	threshold := flag.Int("threshold", 8, "相似度阈值（哈希汉明距离），越小越严格，默认8")
	durationSec := flag.Int("seconds", 8, "用于指纹的音频时长（秒）— 从文件开头读取多少秒用于指纹计算，默认8秒")
	verbose := flag.Bool("v", false, "是否打印详细进度信息")

	flag.Parse()

	if *srcDir == "" || *dstDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	start := time.Now()
	if *verbose {
		log.Printf("开始音频去重：src=%s dst=%s workers=%d threshold=%d seconds=%d\n",
			*srcDir, *dstDir, *workers, *threshold, *durationSec)
	}

	// 1. 扫描文件
	exts := []string{".mp3", ".wav", ".flac", ".aac", ".m4a", ".ogg"} // 支持的扩展
	files, err := scanner.ScanDir(*srcDir, exts)
	if err != nil {
		log.Fatalf("扫描目录失败: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("未在 %s 找到任何支持的音频文件", *srcDir)
	}
	if *verbose {
		log.Printf("扫描到 %d 个音频文件\n", len(files))
	}

	// 2. 并发计算指纹
	type result struct {
		meta dedup.FileMeta
		err  error
	}

	jobs := make(chan string)
	results := make(chan result)
	var wg sync.WaitGroup

	// 启动 worker
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				fp, size, err := fingerprint.FingerprintFromFile(p, *durationSec, 64) // 64-bit 指纹
				r := result{meta: dedup.FileMeta{Path: p, Size: size, FP: fp}, err: err}
				results <- r
			}
		}()
	}

	// 发送任务
	go func() {
		for _, f := range files {
			jobs <- f
		}
		close(jobs)
	}()

	// 收集结果
	var metas []dedup.FileMeta
	var collectErr error
	go func() {
		for res := range results {
			if res.err != nil {
				// 记录第一个错误并继续（不希望单文件失败就中断整个流程）
				if collectErr == nil {
					collectErr = res.err
				}
				log.Printf("警告：处理文件 %s 失败: %v\n", res.meta.Path, res.err)
				continue
			}
			metas = append(metas, res.meta)
			if *verbose {
				log.Printf("指纹计算完成: %s (size=%d bits=%b)\n", res.meta.Path, res.meta.Size, res.meta.FP)
			}
		}
	}()

	// 等待 worker 完成后关闭 results
	wg.Wait()
	close(results)

	if collectErr != nil {
		log.Printf("注意：存在文件处理错误（见上方警告），请核对处理日志")
	}

	if len(metas) == 0 {
		log.Fatalf("没有成功计算任何文件的指纹")
	}

	// 3. 去重（基于汉明距离 + union-find 组建）
	keeps := dedup.SelectKeep(metas, *threshold)

	// 4. 复制保留文件到目标目录
	if err := os.MkdirAll(*dstDir, 0o755); err != nil {
		log.Fatalf("创建目标目录失败: %v", err)
	}
	for _, m := range keeps {
		dstPath := filepath.Join(*dstDir, filepath.Base(m.Path))
		if err := copyutil.CopyFile(m.Path, dstPath); err != nil {
			log.Printf("复制失败: %s -> %s : %v\n", m.Path, dstPath, err)
		} else if *verbose {
			log.Printf("复制成功: %s -> %s\n", m.Path, dstPath)
		}
		reportItems = append(reportItems, report.ReportItem{
			FilePath: m.Path,
			Kept:     true,
			Size:     m.Size,
			NewPath:  dstPath,
		})
	}

	fmt.Printf("完成：源文件 %d，处理成功 %d，保留并复制 %d，耗时 %s\n", len(files), len(metas), len(keeps), time.Since(start))
	if collectErr != nil {
		fmt.Println("注意：存在单文件处理错误，请查看日志。")
	}

	// 处理完成后生成 CSV
	if err := report.WriteCSVReport(reportItems); err != nil {
		fmt.Printf("生成报告失败: %v\n", err)
	}
}
