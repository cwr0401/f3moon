package engine

import (
	"math/rand"

	"github.com/cwr0401/f3moon/internal/model"
)

// NewDeck 生成112张牌
func NewDeck() []*model.Tile {
	var tiles []*model.Tile
	id := 0

	for _, name := range model.AllTileNames {
		color := model.GetTileColor(name)
		numeric := model.NumericValue(name)
		flowerCount := 0
		if model.FlowerTileNames[name] {
			flowerCount = 2 // 乙三五七九 有2张带花
		}
		totalCount := 5 // 每种5张

		for i := 0; i < totalCount; i++ {
			isFlower := i < flowerCount
			tile := &model.Tile{
				ID:       id,
				Name:     name,
				Color:    color,
				IsFlower: isFlower,
				Numeric:  numeric,
			}
			tiles = append(tiles, tile)
			id++
		}
	}

	// 别字: 2张, 红色, 始终带花
	for i := 0; i < 2; i++ {
		tile := &model.Tile{
			ID:       id,
			Name:     model.TileBie,
			Color:    model.ColorRed,
			IsFlower: true,
			Numeric:  0,
		}
		tiles = append(tiles, tile)
		id++
	}

	return tiles
}

// Shuffle 洗牌(Fisher-Yates)
func Shuffle(tiles []*model.Tile, r *rand.Rand) {
	r.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})
}

// Cut 切牌: 从top位置切牌, 将top及以上的牌移到底部
func Cut(tiles []*model.Tile, top int) []*model.Tile {
	if top <= 0 || top >= len(tiles) {
		return tiles
	}
	// tiles[0..top-1] 放到 tiles[top..] 的下面
	result := make([]*model.Tile, len(tiles))
	copy(result, tiles[top:])
	copy(result[len(tiles)-top:], tiles[:top])
	return result
}

// Deal 给3个玩家各发25张牌, 剩余37张为公牌
func Deal(tiles []*model.Tile) (hands [3][]*model.Tile, drawPile []*model.Tile) {
	for i := 0; i < 25; i++ {
		for p := 0; p < 3; p++ {
			hands[p] = append(hands[p], tiles[i*3+p])
		}
	}
	drawPile = tiles[75:]
	return
}
