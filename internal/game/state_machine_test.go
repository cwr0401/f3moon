package game

import (
	"testing"

	"github.com/cwr0401/f3moon/internal/model"
)

// mockNotifier 是一个简单的模拟广播器
type mockNotifier struct {
	messages []NotifyMessage
}

func (m *mockNotifier) Broadcast(msg NotifyMessage) {
	m.messages = append(m.messages, msg)
}

func TestStateMachine_StartGame(t *testing.T) {
	// 创建测试游戏状态
	gs := model.NewGameState("test-room-1", model.GameMode3Player)
	gs.Players[0] = &model.Player{ID: "player-0", Name: "庄家"}
	gs.Players[1] = &model.Player{ID: "player-1", Name: "闲家1"}
	gs.Players[2] = &model.Player{ID: "player-2", Name: "闲家2"}

	notifier := &mockNotifier{}
	sm := NewStateMachine(gs, notifier)

	// 测试开始游戏
	err := sm.StartGame()
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// 验证阶段是否正确转换到切牌阶段
	if gs.Phase != model.PhaseCut {
		t.Errorf("Expected phase %v, got %v", model.PhaseCut, gs.Phase)
	}

	// 验证是否有广播消息
	if len(notifier.messages) == 0 {
		t.Error("Expected broadcast messages")
	}
}

func TestStateMachine_InvalidPhaseTransition(t *testing.T) {
	gs := model.NewGameState("test-room-4", model.GameMode3Player)
	gs.Players[0] = &model.Player{ID: "player-0", Name: "庄家"}
	gs.Players[1] = &model.Player{ID: "player-1", Name: "闲家1"}
	gs.Players[2] = &model.Player{ID: "player-2", Name: "闲家2"}
	gs.Phase = model.PhaseWaiting // 游戏还没开始

	notifier := &mockNotifier{}
	sm := NewStateMachine(gs, notifier)

	// 尝试在等待阶段打牌（应该失败）
	evt := GameEvent{
		Type:     EventDraw,
		PlayerID: "player-0",
	}

	err := sm.HandleEvent(evt)
	if err == nil {
		t.Error("Expected error for invalid phase, got nil")
	}
}

func TestStateMachine_GameEnded(t *testing.T) {
	gs := model.NewGameState("test-room-5", model.GameMode3Player)
	gs.Players[0] = &model.Player{ID: "player-0", Name: "庄家"}
	gs.Players[1] = &model.Player{ID: "player-1", Name: "闲家1"}
	gs.Players[2] = &model.Player{ID: "player-2", Name: "闲家2"}
	gs.Phase = model.PhaseFinished // 游戏已结束

	notifier := &mockNotifier{}
	sm := NewStateMachine(gs, notifier)

	evt := GameEvent{
		Type:     EventDraw,
		PlayerID: "player-0",
	}

	err := sm.HandleEvent(evt)
	if err == nil {
		t.Error("Expected error for finished game, got nil")
	}
}

func TestStateMachine_Game(t *testing.T) {
	gs := model.NewGameState("test-room-6", model.GameMode3Player)
	notifier := &mockNotifier{}
	sm := NewStateMachine(gs, notifier)

	retrieved := sm.Game()
	if retrieved != gs {
		t.Error("Game() should return the same state")
	}
}
