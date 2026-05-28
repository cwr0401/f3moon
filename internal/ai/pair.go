package ai

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// decidePairAction 碰牌决策
func (ai *AIPlayer) decidePairAction(ctx *GameContext) AIDecision {
	if ctx.DiscardTile == nil {
		return AIDecision{Action: model.ActionPass}
	}

	_, pairSize := engine.CanPairWith(ai.player.Hand, ctx.DiscardTile)
	return AIDecision{
		Action:   model.ActionPair,
		PairSize: pairSize,
	}
}
