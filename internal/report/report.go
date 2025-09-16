// 文件路径: internal/report/report.go
package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// ReportItem 表示每个音频文件的处理记录
type ReportItem struct {
	FilePath string // 原始文件路径
	Kept     bool   // 是否保留
	Size     int64  // 文件大小
	NewPath  string // 如果保留，复制到的新路径
}

// WriteCSVReport 将报告写入 CSV 文件
func WriteCSVReport(items []ReportItem) error {
	// 当前目录下生成去重报告，文件名带时间戳
	filename := fmt.Sprintf("audio_dedup_report_%s.csv", time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create report file error: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write([]string{"FilePath", "Kept", "Size", "NewPath"}); err != nil {
		return fmt.Errorf("write csv header error: %w", err)
	}

	// 写入每行记录
	for _, item := range items {
		kept := "No"
		if item.Kept {
			kept = "Yes"
		}
		record := []string{
			item.FilePath,
			kept,
			fmt.Sprintf("%d", item.Size),
			item.NewPath,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write csv record error: %w", err)
		}
	}

	fmt.Printf("去重报告已生成: %s\n", filename)
	return nil
}
