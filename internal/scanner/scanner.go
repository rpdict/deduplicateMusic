// file: internal/scanner/scanner.go
// package: scanner
//
// 提供目录扫描功能，按扩展名筛选音频文件。
package scanner

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// ScanDir 扫描 root 目录，返回匹配 exts 中扩展名（小写）的文件路径列表。
// exts 样例：[]string{".mp3", ".wav"}
func ScanDir(root string, exts []string) ([]string, error) {
	if len(exts) == 0 {
		return nil, nil
	}
	extMap := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		extMap[strings.ToLower(e)] = struct{}{}
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 如果单路径访问错误，继续其他路径
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := extMap[ext]; ok {
			// 确认为常见音频扩展
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
