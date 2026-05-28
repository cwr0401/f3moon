package util

import (
	"math/rand"
	"time"
)

// NewSeededRand 创建带种子的随机数生成器
func NewSeededRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandInt 返回[min, max)之间的随机整数
func RandInt(r *rand.Rand, min, max int) int {
	return r.Intn(max-min) + min
}

// ShuffleIntSlice 打乱整数切片
func ShuffleIntSlice(r *rand.Rand, slice []int) {
	r.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}
