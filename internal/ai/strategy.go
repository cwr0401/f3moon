package ai

import "github.com/cwr0401/f3moon/internal/model"

// shouldPair 是否应该碰牌(基于strategy.md的对牌策略)
func (ai *AIPlayer) shouldPair(ctx *GameContext) bool {
	totalHu := ai.totalHu()
	tile := ctx.DiscardTile

	if tile == nil {
		return false
	}

	// 别字不能碰
	if tile.Name == model.TileBie {
		return false
	}

	switch ai.difficulty {
	case Easy:
		return true // 简单模式: 能碰就碰
	case Medium:
		// 胡数>=17: 当碰则碰
		if totalHu >= 17 {
			return true
		}
		// 胡数>=13: 碰红色牌或带花黑色牌
		if totalHu >= 13 {
			if tile.Color == model.ColorRed || tile.IsFlower {
				return true
			}
		}
		// 胡数>=14: 都可以碰
		if totalHu >= 14 {
			return true
		}
		return false
	case Hard:
		// 完整策略
		if totalHu >= 17 {
			return true
		}
		// 经牌策略: 花经选择碰, 白经不碰
		if model.IsJingName(tile.Name) {
			if tile.IsFlower {
				return true
			}
			return false
		}
		// 红色牌/带花黑色牌优先碰
		if tile.Color == model.ColorRed || tile.IsFlower {
			return totalHu >= 13
		}
		return totalHu >= 14
	}
	return false
}

// shouldGanta 是否应该赶塔
func (ai *AIPlayer) shouldGanta(ctx *GameContext) bool {
	// 赶塔增加胡数, 一般都应该赶
	if ai.difficulty == Easy {
		return true
	}
	// 困难模式: 评估赶塔后是否影响听牌
	totalHu := ai.totalHu()
	return totalHu+20 >= 17 || totalHu >= 10 // 赶塔后胡数显著增加
}

// shouldPlaySafe 是否应该防守(打跟张)
func (ai *AIPlayer) shouldPlaySafe(ctx *GameContext) bool {
	if ai.difficulty == Easy {
		return false
	}

	totalHu := ai.totalHu()

	// 胡数不足10, 基本无和牌希望, 打跟张
	if totalHu < 10 {
		return true
	}

	// 检查是否有玩家多次统牌(可能大胡)
	tongCount := 0
	for _, player := range ctx.Game.Players {
		if player == nil || player.ID == ai.player.ID {
			continue
		}
		for _, comb := range player.OpenCombs {
			if comb.Type == model.CombSingle && len(comb.Tiles) >= 4 {
				tongCount++
			}
		}
	}
	return tongCount >= 2 // 有人多次统牌, 打防守
}
