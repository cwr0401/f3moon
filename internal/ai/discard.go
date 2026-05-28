package ai

import (
	"sort"

	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// DiscardCandidate 出牌候选
type DiscardCandidate struct {
	Tile  *model.Tile
	Score float64 // 分越低越优先打出
}

// decideDiscard 出牌决策
func (ai *AIPlayer) decideDiscard(ctx *GameContext) AIDecision {
	candidates := ai.evaluateDiscardCandidates(ctx)

	if len(candidates) == 0 {
		// 没有候选, 打出第一张
		if len(ai.player.Hand) > 0 {
			return AIDecision{Action: model.ActionDiscard, TileID: ai.player.Hand[0].ID}
		}
		return AIDecision{Action: model.ActionPass}
	}

	// 按评分排序(分越低越优先打出)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score < candidates[j].Score
	})

	// 跟张策略: 如果判断需防守, 优先打出别人打过的牌
	if ai.shouldPlaySafe(ctx) {
		if safe := ai.findSafeDiscard(ctx, candidates); safe != nil {
			return AIDecision{Action: model.ActionDiscard, TileID: safe.ID}
		}
	}

	return AIDecision{Action: model.ActionDiscard, TileID: candidates[0].Tile.ID}
}

// evaluateDiscardCandidates 评估每张牌的出牌价值
func (ai *AIPlayer) evaluateDiscardCandidates(ctx *GameContext) []DiscardCandidate {
	var candidates []DiscardCandidate
	hand := ai.player.Hand
	dangJing := ai.player.DangJing

	for _, tile := range hand {
		// 别字绝对不打
		if tile.Name == model.TileBie {
			continue
		}

		score := 0.0

		switch ai.difficulty {
		case Easy:
			score = ai.easyDiscardScore(tile, hand, dangJing)
		case Medium:
			score = ai.mediumDiscardScore(tile, hand, dangJing)
		case Hard:
			score = ai.hardDiscardScore(tile, hand, dangJing, ctx)
		}

		candidates = append(candidates, DiscardCandidate{Tile: tile, Score: score})
	}

	return candidates
}

// easyDiscardScore 简单模式评分
func (ai *AIPlayer) easyDiscardScore(tile *model.Tile, hand []*model.Tile, dangJing model.TileName) float64 {
	score := 0.0
	// 经牌权重
	if model.IsJingName(tile.Name) {
		score += 20
	}
	// 带花权重
	if tile.IsFlower {
		score += 5
	}
	// 红色权重
	if tile.Color == model.ColorRed {
		score += 3
	}
	// 参与组合的权重
	score += float64(ai.countCombInvolvement(tile, hand, dangJing)) * 10
	return score
}

// mediumDiscardScore 中等模式评分
func (ai *AIPlayer) mediumDiscardScore(tile *model.Tile, hand []*model.Tile, dangJing model.TileName) float64 {
	score := ai.easyDiscardScore(tile, hand, dangJing)

	// 胡数贡献
	huBefore := ai.totalHu()
	testHand := removeTileFromSlice(hand, tile.ID)
	arrangement := engine.BestArrangement(testHand, dangJing)
	testHu := arrangement.TotalHu
	for _, comb := range ai.player.OpenCombs {
		testHu += engine.CalcOpenCombHu(comb, dangJing)
	}
	score += float64(huBefore-testHu) * 5

	return score
}

// hardDiscardScore 困难模式评分
func (ai *AIPlayer) hardDiscardScore(tile *model.Tile, hand []*model.Tile, dangJing model.TileName, ctx *GameContext) float64 {
	score := ai.mediumDiscardScore(tile, hand, dangJing)

	// 听牌影响
	testHand := removeTileFromSlice(hand, tile.ID)
	tingResults := engine.DetectTing(testHand, ai.player.OpenCombs, dangJing)
	if len(tingResults) > 0 {
		// 打出后仍然能听牌, 增加权重(不要打出影响听牌的牌)
		bestTingCount := 0
		for _, tr := range tingResults {
			if len(tr.TingTiles) > bestTingCount {
				bestTingCount = len(tr.TingTiles)
			}
		}
		score += float64(bestTingCount) * 15
	}

	// 剩余张数影响(打出剩余多的牌更安全)
	remaining := ai.memory.RemainingCount(tile.Name)
	score -= float64(remaining) * 2

	return score
}

// countCombInvolvement 计算一张牌参与的组合数
func (ai *AIPlayer) countCombInvolvement(tile *model.Tile, hand []*model.Tile, dangJing model.TileName) int {
	count := 0
	arrangement := engine.BestArrangement(hand, dangJing)
	tileID := tile.ID

	for _, comb := range arrangement.Complete {
		for _, t := range comb.Tiles {
			if t.ID == tileID {
				count++
				break
			}
		}
	}
	for _, comb := range arrangement.Incomplete {
		for _, t := range comb.Tiles {
			if t.ID == tileID {
				count++
				break
			}
		}
	}
	return count
}

// findSafeDiscard 找到安全的出牌(跟张)
func (ai *AIPlayer) findSafeDiscard(ctx *GameContext, candidates []DiscardCandidate) *model.Tile {
	// 优先找最近别人打过的同名牌
	for _, c := range candidates {
		if ai.memory.IsSafeDiscard(c.Tile.Name) {
			return c.Tile
		}
	}
	return nil
}

// removeTileFromSlice 从牌切片中移除指定ID的牌
func removeTileFromSlice(tiles []*model.Tile, tileID int) []*model.Tile {
	result := make([]*model.Tile, 0, len(tiles))
	for _, t := range tiles {
		if t.ID != tileID {
			result = append(result, t)
		}
	}
	return result
}
