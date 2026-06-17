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

// MinCutPosition 切牌位置下界 (rules.md:121 严格大于 37)
const MinCutPosition = 38

// MaxCutPosition 切牌位置上界 (rules.md:121 严格小于 111)
const MaxCutPosition = 110

// ValidateCutPosition 校验切牌位置是否合法.
// 规则 rules.md:121: 切牌位置必须 >37 且 <111, 即合法区间 [38, 110].
func ValidateCutPosition(position int) bool {
	return position >= MinCutPosition && position <= MaxCutPosition
}

// CutIDs 切牌: 从 top 位置切牌, 将 top 及以上的牌移到底部.
// 规则: 仅允许 top ∈ [38, 110]; 超出范围时返回原数组不做修改 (调用方应预先校验).
func CutIDs(ids []uint8, top int) []uint8 {
	if !ValidateCutPosition(top) || top >= len(ids) {
		return ids
	}
	result := make([]uint8, len(ids))
	copy(result, ids[top:])
	copy(result[len(ids)-top:], ids[:top])
	return result
}

// Cut 切牌 (保持向后兼容).
// 同样使用 ValidateCutPosition 做边界校验.
func Cut(tiles []*model.Tile, top int) []*model.Tile {
	if !ValidateCutPosition(top) || top >= len(tiles) {
		return tiles
	}
	result := make([]*model.Tile, len(tiles))
	copy(result, tiles[top:])
	copy(result[len(tiles)-top:], tiles[:top])
	return result
}

// DealIDs 给3个玩家各发25张牌, 剩余37张为公牌 (基于 ID 数组).
// 规则 rules.md:131-135: 依次给庄家(0)、闲一(1)、闲二(2)各发1张, 循环25轮 (轮转/round-robin).
// 与 handler/game.go::Deal 的发牌顺序保持一致, 公牌从索引 75 开始.
func DealIDs(ids []uint8) (hands [3][]uint8, drawPile []uint8) {
	for round := 0; round < 25; round++ {
		for p := 0; p < 3; p++ {
			hands[p] = append(hands[p], ids[round*3+p])
		}
	}
	drawPile = ids[75:]
	return
}

// Deal 发牌 (保持向后兼容). 与 DealIDs 相同的轮转语义.
func Deal(tiles []*model.Tile) (hands [3][]*model.Tile, drawPile []*model.Tile) {
	for round := 0; round < 25; round++ {
		for p := 0; p < 3; p++ {
			hands[p] = append(hands[p], tiles[round*3+p])
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
