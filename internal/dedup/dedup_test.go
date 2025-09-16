// file: internal/dedup/dedup_test.go
// package: dedup
//
// 测试去重逻辑：构造几个 FileMeta，其中两个具有相同指纹，应保留体积更大的那个。
package dedup

import (
	"testing"
)

func TestSelectKeepBasic(t *testing.T) {
	// 构造 3 个文件：0 和 1 相似，2 不相似
	files := []FileMeta{
		{Path: "a.mp3", Size: 1000, FP: 0x0f0f0f0f0f0f0f0f},
		{Path: "b.mp3", Size: 2000, FP: 0x0f0f0f0f0f0f0f0f}, // 与 a 相似且体积更大 -> 应保留 b
		{Path: "c.mp3", Size: 1500, FP: 0xf0f0f0f0f0f0f0f0}, // 不相似 -> 单独保留
	}
	keeps := SelectKeep(files, 4)
	if len(keeps) != 2 {
		t.Fatalf("期望保留 2 个文件，实际 %d", len(keeps))
	}
	// 检查 b.mp3 与 c.mp3 被保留
	paths := map[string]bool{}
	for _, k := range keeps {
		paths[k.Path] = true
	}
	if !paths["b.mp3"] || !paths["c.mp3"] {
		t.Fatalf("保留文件不正确: %#v", keeps)
	}
}
