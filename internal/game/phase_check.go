package game

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// WinResult 和牌查验结果
type WinResult struct {
	Winner      int                  `json:"winner"`
	WinType     model.WinType        `json:"win_type"`
	TotalHu     int                  `json:"total_hu"`
	Po          int                  `json:"po"`
	DangJing    model.TileName       `json:"dang_jing"`
	HandCombs   []*model.Combination `json:"hand_combs"`
	OpenCombs   []*model.Combination `json:"open_combs"`
}

// CheckWin 执行和牌查验
func (sm *StateMachine) CheckWin() *WinResult {
	if sm.game.Winner < 0 {
		return nil
	}

	player := sm.game.Players[sm.game.Winner]
	dangJing := player.DangJing

	// 理牌
	arrangement := engine.BestArrangement(player.Hand, dangJing)

	handCombs := make([]*model.Combination, 0)
	handCombs = append(handCombs, arrangement.Complete...)
	handCombs = append(handCombs, arrangement.Incomplete...)

	// 计算总胡数
	totalHu := engine.CalcTotalHu(handCombs, player.OpenCombs, dangJing)

	result := &WinResult{
		Winner:    sm.game.Winner,
		WinType:   sm.game.WinType,
		TotalHu:   totalHu,
		Po:        engine.CalcPo(totalHu),
		DangJing:  dangJing,
		HandCombs: handCombs,
		OpenCombs: player.OpenCombs,
	}

	// 广播查验结果
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyCheck,
		GameID: sm.game.ID,
		Data:   result,
	})

	// 进入结束阶段
	sm.game.Phase = model.PhaseFinished

	return result
}
