package ai

import (
	"github.com/cwr0401/f3moon/internal/model"
)

// Memory AI玩家的出牌记忆
type Memory struct {
	discards   map[model.TileName]int  // 出牌区中各牌的数量
	openTiles  map[model.TileName]int  // 明牌区中各牌的数量
}

// NewMemory 创建记忆
func NewMemory() *Memory {
	return &Memory{
		discards:  make(map[model.TileName]int),
		openTiles: make(map[model.TileName]int),
	}
}

// RecordDiscard 记录出牌
func (m *Memory) RecordDiscard(tile *model.Tile) {
	m.discards[tile.Name]++
}

// RecordOpenTile 记录明牌
func (m *Memory) RecordOpenTile(tile *model.Tile) {
	m.openTiles[tile.Name]++
}

// RemainingCount 计算某牌的剩余数量(不出现在出牌区+明牌区的数量)
// 总数: 普通牌5张, 别字2张
func (m *Memory) RemainingCount(name model.TileName) int {
	total := 5
	if name == model.TileBie {
		total = 2
	}
	return total - m.discards[name] - m.openTiles[name]
}

// RefreshFromGame 从游戏状态刷新记忆
func (m *Memory) RefreshFromGame(game *model.GameState) {
	m.discards = make(map[model.TileName]int)
	m.openTiles = make(map[model.TileName]int)

	for _, tile := range game.DiscardPile {
		m.discards[tile.Name]++
	}
	for _, player := range game.Players {
		if player == nil {
			continue
		}
		for _, tile := range player.OpenTiles {
			m.openTiles[tile.Name]++
		}
	}
}

// IsSafeDiscard 判断打出某牌是否安全(该牌已在出牌区出现过)
func (m *Memory) IsSafeDiscard(name model.TileName) bool {
	return m.discards[name] > 0
}
