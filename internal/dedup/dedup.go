// file: internal/dedup/dedup.go
// package: dedup
//
// 去重核心逻辑：
//   - 数据结构 FileMeta 保存文件路径、大小、指纹。
//   - 使用 union-find（并查集）把“相似”文件（汉明距离 <= threshold）连成组件。
//   - 对每个组件选择文件大小最大的作为保留（如果大小相同则按路径字典序保留第一个）。
package dedup

import (
	"deduplicateMusic/internal/fingerprint"
	"sort"
	"sync"
)

// FileMeta 表示已计算指纹的文件信息
type FileMeta struct {
	Path string
	Size int64
	FP   uint64
}

// SelectKeep 接受文件列表与阈值（汉明距离），返回保留的文件列表。
// 算法：对每对文件比较，若汉明距离 <= threshold 则 union(i,j)；最后对每个并查集选择最大文件。
func SelectKeep(files []FileMeta, threshold int) []FileMeta {
	n := len(files)
	if n == 0 {
		return nil
	}
	uf := newUnionFind(n)

	// 并行比较所有对（简单的 N^2；对于数千文件可能慢，可进一步分桶优化）
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := i + 1; j < n; j++ {
				dist := fingerprint.HammingDistance(files[i].FP, files[j].FP)
				if dist <= threshold {
					uf.union(i, j)
				}
			}
		}()
	}
	wg.Wait()

	// group by root
	groups := make(map[int][]int)
	for i := 0; i < n; i++ {
		r := uf.find(i)
		groups[r] = append(groups[r], i)
	}

	// 选出每组中 size 最大的文件
	var keeps []FileMeta
	for _, idxs := range groups {
		// 找最大 size，否则按字典序最小
		sort.Slice(idxs, func(i, j int) bool {
			a, b := files[idxs[i]], files[idxs[j]]
			if a.Size != b.Size {
				return a.Size > b.Size // 降序，方便取第0个
			}
			return a.Path < b.Path
		})
		keeps = append(keeps, files[idxs[0]])
	}

	// 可选：按路径排序返回（方便查看）
	sort.Slice(keeps, func(i, j int) bool { return keeps[i].Path < keeps[j].Path })
	return keeps
}

// ----------------- 并查集实现 -----------------
type unionFind struct {
	parent []int
	rank   []int
}

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	r := make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = i
		r[i] = 0
	}
	return &unionFind{parent: p, rank: r}
}

func (u *unionFind) find(x int) int {
	if u.parent[x] != x {
		u.parent[x] = u.find(u.parent[x])
	}
	return u.parent[x]
}

func (u *unionFind) union(a, b int) {
	ar := u.find(a)
	br := u.find(b)
	if ar == br {
		return
	}
	if u.rank[ar] < u.rank[br] {
		u.parent[ar] = br
	} else if u.rank[ar] > u.rank[br] {
		u.parent[br] = ar
	} else {
		u.parent[br] = ar
		u.rank[ar]++
	}
}
