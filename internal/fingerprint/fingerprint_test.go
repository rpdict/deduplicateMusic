// file: internal/fingerprint/fingerprint_test.go
// package: fingerprint
//
// 测试指纹与汉明距离的基础行为，使用合成样本避免依赖 ffmpeg。
package fingerprint

import (
	"testing"
)

func makeConstantSamples(val int16, n int) []int16 {
	s := make([]int16, n)
	for i := 0; i < n; i++ {
		s[i] = val
	}
	return s
}

func TestFingerprintFromSamplesConsistency(t *testing.T) {
	s1 := makeConstantSamples(1000, 8000)
	s2 := makeConstantSamples(1005, 8000) // 与 s1 很接近（小幅偏移）
	fp1 := FingerprintFromSamples(s1, 64)
	fp2 := FingerprintFromSamples(s2, 64)
	dist := HammingDistance(fp1, fp2)
	if dist > 6 { // 经验值：非常相近的样本，汉明距离应很小
		t.Fatalf("期望相似样本的汉明距离较小，但得到了 %d", dist)
	}
}

func TestHammingDistanceBasic(t *testing.T) {
	var a uint64 = 0b101010
	var b uint64 = 0b111000
	dist := HammingDistance(a, b)
	// 101010 xor 111000 = 010010 -> 2 个位不同
	if dist != 2 {
		t.Fatalf("期望汉明距离 2，实际 %d", dist)
	}
}
