package game

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// StateMachine 游戏状态机
type StateMachine struct {
	mu        sync.RWMutex
	game      *model.GameState
	broadcast NotifyBroadcaster
	rng       *rand.Rand
}

// NewStateMachine 创建状态机
func NewStateMachine(game *model.GameState, broadcast NotifyBroadcaster) *StateMachine {
	return &StateMachine{
		game:      game,
		broadcast: broadcast,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Game 返回当前游戏状态
func (sm *StateMachine) Game() *model.GameState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.game
}

// HandleEvent 处理游戏事件
func (sm *StateMachine) HandleEvent(evt GameEvent) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	switch sm.game.Phase {
	case model.PhaseCut:
		return sm.handleCut(evt)
	case model.PhaseDeal:
		return errors.New("dealing in progress")
	case model.PhaseTongAsk:
		return sm.handleTongAsk(evt)
	case model.PhasePlay:
		return sm.handlePlay(evt)
	case model.PhaseWaiting, model.PhaseShuffle:
		return errors.New("game not started")
	case model.PhaseCheck, model.PhaseFinished:
		return errors.New("game already ended")
	}
	return nil
}

// StartGame 开始游戏
func (sm *StateMachine) StartGame() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 如果还没有牌栈，则洗牌（兼容旧代码）
	if sm.game.DrawPile == nil || len(sm.game.DrawPile) == 0 {
		deck := engine.NewDeck()
		engine.Shuffle(deck, sm.rng)
		sm.game.DrawPile = deck
	}

	sm.game.Phase = model.PhaseCut

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyCutWait,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"cut_player_index": sm.game.CutPlayer(),
		},
	})

	return nil
}

// handleCut 处理切牌
func (sm *StateMachine) handleCut(evt GameEvent) error {
	if evt.Type != EventCut {
		return errors.New("invalid event type for cut phase")
	}
	data, ok := evt.Data.(CutData)
	if !ok {
		return errors.New("invalid cut data")
	}

	// 校验切牌人身份
	cutIdx := sm.game.CutPlayer()
	cutPlayer := sm.game.Players[cutIdx]
	if cutPlayer == nil || cutPlayer.ID != evt.PlayerID {
		return errors.New("not the cut player")
	}

	position := data.Position
	if position < 37 || position > 110 {
		return errors.New("cut position out of range [37, 110]")
	}

	// 切牌
	sm.game.DrawPile = engine.Cut(sm.game.DrawPile, position)

	// 异步起牌(phase 转换由 dealCards 内部完成)
	go sm.dealCards()

	return nil
}

// handleTongAsk 处理请统
func (sm *StateMachine) handleTongAsk(evt GameEvent) error {
	if evt.Type != EventTong {
		return nil
	}
	return sm.processTongDecision(evt)
}

// handlePlay 处理打牌阶段事件
func (sm *StateMachine) handlePlay(evt GameEvent) error {
	switch evt.Type {
	case EventDraw:
		return sm.handleDraw(evt)
	case EventDiscard:
		return sm.handleDiscard(evt)
	case EventPair:
		return sm.handlePair(evt)
	case EventGanta:
		return sm.handleGanta(evt)
	case EventWin:
		return sm.handleWin(evt)
	case EventPass:
		return sm.handlePass(evt)
	case EventDangJing:
		return sm.handleDangJing(evt)
	}
	return nil
}

// broadcastState 广播状态更新
func (sm *StateMachine) broadcastState() {
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyPhaseChange,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"phase":        sm.game.Phase,
			"current_turn": sm.game.CurrentTurn,
			"draw_pile_size": sm.game.DrawPileSize(),
		},
	})
}

// broadcastPlayerHand 向特定玩家推送手牌
func (sm *StateMachine) broadcastPlayerHand(playerIdx int) {
	player := sm.game.Players[playerIdx]
	if player == nil {
		return
	}
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyPlayerHand,
		GameID: sm.game.ID,
		Player: player.ID,
		Data: map[string]interface{}{
			"hand":    player.Hand,
			"dang_jing": player.DangJing,
		},
	})
}
