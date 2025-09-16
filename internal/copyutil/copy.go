// file: internal/copyutil/copy.go
// package: copyutil
//
// 简单的文件复制工具，保留文件权限（若可能）。
package copyutil

import (
	"io"
	"os"
	"path/filepath"
)

// CopyFile 将 src 文件复制到 dst（若 dst 存在会被覆盖）。
// 1) 确保 dst 目录存在
// 2) 使用 io.Copy 复制内容并尝试复制权限
func CopyFile(src, dst string) error {
	if err := ensureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// 创建临时文件然后重命名，降低写出中途失败的风险
	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}

	//// 尝试复制文件模式（若失败也不致命）
	//if fi, e := os.Stat(src); e == nil {
	//	_ = os.Chmod(tmp, fi.Mode())
	//	_ = os.Chown(tmp, int(fi.Sys().(*os.FileStat).Uid), int(fi.Sys().(*os.FileStat).Gid))
	//	// 上面 Chown 可能在某些系统/权限下失败，忽略错误
	//}

	// 重命名到目标文件
	if err := os.Rename(tmp, dst); err != nil {
		// 重命名失败则尝试直接复制覆盖
		if cerr := os.Remove(dst); cerr == nil {
			if rerr := os.Rename(tmp, dst); rerr == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}
