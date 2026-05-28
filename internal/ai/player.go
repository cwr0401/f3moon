package ai

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// Difficulty AI难度
type Difficulty int

const (
	Easy   Difficulty = iota // 简单: 随机决策
	Medium                   // 中等: 基本策略
	Hard                     // 困难: 完整策略
)

// AIDecision AI决策结果
type AIDecision struct {
	Action   model.TurnAction `json:"action"`
	TileID   int              `json:"tile_id"`    // 出牌/碰牌对应的牌ID
	TileName model.TileName   `json:"tile_name"`  // 统牌对应的牌名
	TongSize int              `json:"tong_size"`   // 统牌大小(4或5)
	DangJing model.TileName   `json:"dang_jing"`   // 当经选择
	PairSize int              `json:"pair_size"`   // 碰牌大小(3=对,4=招,5=泛)
}

// AIPlayer AI玩家
type AIPlayer struct {
	player     *model.Player
	difficulty Difficulty
	memory     *Memory
}

// NewAIPlayer 创建AI玩家
func NewAIPlayer(player *model.Player, difficulty Difficulty) *AIPlayer {
	return &AIPlayer{
		player:     player,
		difficulty: difficulty,
		memory:     NewMemory(),
	}
}

// GameContext 游戏上下文
type GameContext struct {
	Game        *model.GameState
	PlayerIdx   int
	DiscardTile *model.Tile // 别人打出的牌(碰牌/和牌判定时)
}

// Decide 做出决策
func (ai *AIPlayer) Decide(ctx *GameContext) AIDecision {
	player := ai.player
	hand := player.Hand
	_ = hand // used in sub-functions

	// 1. 检查是否可以和牌
	if ai.canWin(ctx) {
		return AIDecision{Action: model.ActionWin}
	}

	// 2. 选择当经(如果还没选)
	if player.DangJing == "" {
		jing := engine.ChooseDangJing(hand, player.OpenCombs)
		return AIDecision{Action: model.ActionDangJing, DangJing: jing}
	}

	// 3. 是否碰牌
	if ai.canPair(ctx) {
		if ai.shouldPair(ctx) {
			return ai.decidePairAction(ctx)
		}
		return AIDecision{Action: model.ActionPass}
	}

	// 4. 是否赶塔
	if ai.canGanta(ctx) {
		if ai.shouldGanta(ctx) {
			return AIDecision{Action: model.ActionGanta}
		}
	}

	// 5. 是否统牌
	if ai.canTong(ctx) {
		if ai.shouldTong(ctx) {
			return ai.decideTongAction(ctx)
		}
	}

	// 6. 出牌决策
	return ai.decideDiscard(ctx)
}

// canWin 是否可以和牌
func (ai *AIPlayer) canWin(ctx *GameContext) bool {
	if ctx.DiscardTile != nil {
		return engine.CheckDiscardWin(ai.player.Hand, ctx.DiscardTile, ai.player.OpenCombs, ai.player.DangJing)
	}
	return engine.CheckSelfWin(ai.player.Hand, ai.player.OpenCombs, ai.player.DangJing)
}

// canPair 是否可以碰牌
func (ai *AIPlayer) canPair(ctx *GameContext) bool {
	if ctx.DiscardTile == nil || !ai.player.CanPair() {
		return false
	}
	can, _ := engine.CanPairWith(ai.player.Hand, ctx.DiscardTile)
	return can
}

// canGanta 是否可以赶塔
func (ai *AIPlayer) canGanta(ctx *GameContext) bool {
	// 简化: 检查手牌最后一张(刚起的牌)是否与明牌区统相同
	if len(ai.player.Hand) == 0 {
		return false
	}
	lastTile := ai.player.Hand[len(ai.player.Hand)-1]
	return engine.CanGanta(ai.player, lastTile)
}

// canTong 是否可以统牌
func (ai *AIPlayer) canTong(ctx *GameContext) bool {
	can, _ := engine.CanTongWithHand(ai.player.Hand)
	return can
}

// totalHu 计算当前总胡数
func (ai *AIPlayer) totalHu() int {
	arrangement := engine.BestArrangement(ai.player.Hand, ai.player.DangJing)
	totalHu := arrangement.TotalHu
	for _, comb := range ai.player.OpenCombs {
		totalHu += engine.CalcOpenCombHu(comb, ai.player.DangJing)
	}
	return totalHu
}
