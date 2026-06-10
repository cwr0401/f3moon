package engine

import (
	"math/rand"
	"time"

	"github.com/cwr0401/f3moon/internal/model"
)

// NewDeck 生成112张牌 (保持向后兼容)
func NewDeck() []*model.Tile {
	return model.GetTilesFromIDs(model.GetAllTileIDs())
}

// NewDeckIDs 返回有序的牌 ID 数组 (0-111)
func NewDeckIDs() []uint8 {
	return model.GetAllTileIDs()
}

// ShuffleIDs 洗牌 (Fisher-Yates) - 仅打乱 ID 数组
func ShuffleIDs(ids []uint8, r *rand.Rand) {
	r.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})
}

// Shuffle 洗牌 (保持向后兼容)
func Shuffle(tiles []*model.Tile, r *rand.Rand) {
	r.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})
}

// CutIDs 切牌: 从 top 位置切牌, 将 top 及以上的牌移到底部
func CutIDs(ids []uint8, top int) []uint8 {
	if top <= 0 || top >= len(ids) {
		return ids
	}
	result := make([]uint8, len(ids))
	copy(result, ids[top:])
	copy(result[len(ids)-top:], ids[:top])
	return result
}

// Cut 切牌 (保持向后兼容)
func Cut(tiles []*model.Tile, top int) []*model.Tile {
	if top <= 0 || top >= len(tiles) {
		return tiles
	}
	result := make([]*model.Tile, len(tiles))
	copy(result, tiles[top:])
	copy(result[len(tiles)-top:], tiles[:top])
	return result
}

// DealIDs 给3个玩家各发25张牌, 剩余37张为公牌 (基于 ID 数组)
func DealIDs(ids []uint8) (hands [3][]uint8, drawPile []uint8) {
	for i := 0; i < 25; i++ {
		for p := 0; p < 3; p++ {
			hands[p] = append(hands[p], ids[i*3+p])
		}
	}
	drawPile = ids[75:]
	return
}

// Deal 发牌 (保持向后兼容)
func Deal(tiles []*model.Tile) (hands [3][]*model.Tile, drawPile []*model.Tile) {
	for i := 0; i < 25; i++ {
		for p := 0; p < 3; p++ {
			hands[p] = append(hands[p], tiles[i*3+p])
		}
	}
	drawPile = tiles[75:]
	return
}

// NewShuffledDeckStack 创建一个打乱后的牌 ID 栈
// 使用指定种子（或当前时间）随机打乱
// 返回的切片索引 0 为栈顶，索引 111 为栈底
func NewShuffledDeckStack(seed ...int64) []uint8 {
	ids := model.GetAllTileIDs()

	var s int64
	if len(seed) > 0 {
		s = seed[0]
	} else {
		s = time.Now().UnixNano()
	}

	r := rand.New(rand.NewSource(s))
	ShuffleIDs(ids, r)

	return ids
}

// NewShuffledDeck 类似 NewShuffledDeckStack，但返回 *Tile 切片
func NewShuffledDeck(seed ...int64) []*model.Tile {
	ids := NewShuffledDeckStack(seed...)
	return model.GetTilesFromIDs(ids)
}
