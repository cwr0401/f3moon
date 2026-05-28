package ai

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// shouldTong 是否应该统牌
func (ai *AIPlayer) shouldTong(ctx *GameContext) bool {
	can, _ := engine.CanTongWithHand(ai.player.Hand)
	if !can {
		return false
	}
	if ai.difficulty == Easy {
		return true
	}
	return ai.totalHu() >= 17 || ai.totalHu() >= 10
}

// decideTongAction 统牌决策
func (ai *AIPlayer) decideTongAction(ctx *GameContext) AIDecision {
	_, tongNames := engine.CanTongWithHand(ai.player.Hand)

	if len(tongNames) == 0 {
		return AIDecision{Action: model.ActionPass}
	}

	// 选择最优统牌(利益最大化)
	bestName := tongNames[0]
	bestHu := -1

	for _, name := range tongNames {
		// 模拟统牌后的胡数
		tongSize := 4
		if ai.player.HasTileInHand(name) >= 5 {
			tongSize = 5
		}

		// 统牌增加的胡数
		simHu := ai.simulateTongHu(name, tongSize)
		if simHu > bestHu {
			bestHu = simHu
			bestName = name
		}
	}

	// 判断统牌是否有利于和牌
	tongSize := 4
	if ai.player.HasTileInHand(bestName) >= 5 {
		tongSize = 5
	}

	// 策略: 胡数>=17时尽量统(追求大胡), 否则评估是否影响听牌
	if ai.totalHu() >= 17 || ai.difficulty == Easy {
		return AIDecision{
			Action:   model.ActionTong,
			TileName: bestName,
			TongSize: tongSize,
		}
	}

	// 中等/困难模式: 评估统牌是否会让手牌更完整
	if ai.difficulty >= Medium {
		// 如果不统牌, 这些牌能形成更好的组合, 则不统
		hand := ai.player.Hand
		arrangement := engine.BestArrangement(hand, ai.player.DangJing)

		// 如果当前排列已经有完整组合覆盖该牌, 不统
		for _, comb := range arrangement.Complete {
			for _, t := range comb.Tiles {
				if t.Name == bestName {
					// 该牌已参与完整组合, 统牌可能破坏组合
					if ai.totalHu() < 17 {
						return AIDecision{Action: model.ActionPass}
					}
				}
			}
		}
	}

	return AIDecision{
		Action:   model.ActionTong,
		TileName: bestName,
		TongSize: tongSize,
	}
}

// simulateTongHu 模拟统牌后的胡数
func (ai *AIPlayer) simulateTongHu(name model.TileName, tongSize int) int {
	// 创建统牌组合
	var tongTiles []*model.Tile
	remaining := make([]*model.Tile, 0)
	count := 0

	for _, t := range ai.player.Hand {
		if t.Name == name && count < tongSize {
			tongTiles = append(tongTiles, t)
			count++
		} else {
			remaining = append(remaining, t)
		}
	}

	// 模拟统牌后的手牌 + 明牌区
	arrangement := engine.BestArrangement(remaining, ai.player.DangJing)
	hu := arrangement.TotalHu

	// 统牌组合的胡数
	comb := model.NewCombination(model.CombSingle, model.CombComplete, tongTiles)
	hu += engine.CalcCombHu(comb, ai.player.DangJing)

	// 加上已有明牌区胡数
	for _, c := range ai.player.OpenCombs {
		hu += engine.CalcOpenCombHu(c, ai.player.DangJing)
	}

	return hu
}
