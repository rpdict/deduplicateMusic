// file: internal/scanner/scanner_test.go
// package: scanner
//
// 测试 ScanDir：创建临时目录并生成带不同扩展名的文件，确保只返回预期扩展的文件。
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirBasic(t *testing.T) {
	td := t.TempDir()
	// 创建样例文件
	files := []string{"a.mp3", "b.wav", "c.txt", "d.FLAC"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(td, f), []byte("dummy"), 0o644); err != nil {
			t.Fatalf("写临时文件失败: %v", err)
		}
	}
	exts := []string{".mp3", ".wav", ".flac"}
	found, err := ScanDir(td, exts)
	if err != nil {
		t.Fatalf("ScanDir 错误: %v", err)
	}
	// 期望找到 a.mp3, b.wav, d.FLAC -> 共 3 个（扩展名大小写不敏感）
	if len(found) != 3 {
		t.Fatalf("期望 3 个文件，实际 %d: %#v", len(found), found)
	}
}
