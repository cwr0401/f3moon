package game

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// dealCards 起牌: 三人各起25张
func (sm *StateMachine) dealCards() {
	for i := 0; i < 25; i++ {
		for p := 0; p < 3; p++ {
			tile := sm.game.DrawFromTop()
			if tile != nil {
				sm.game.Players[p].AddTileToHand(tile)
			}
		}
	}

	// 通知各玩家手牌
	for i := 0; i < 3; i++ {
		sm.broadcastPlayerHand(i)
	}

	// 庄家起第26张牌
	extraTile := sm.game.DrawFromTop()
	if extraTile != nil {
		sm.game.Players[0].AddTileToHand(extraTile)
	}

	// 选择当经
	for i := 0; i < 3; i++ {
		player := sm.game.Players[i]
		if player.DangJing == "" {
			player.DangJing = engine.ChooseDangJing(player.Hand, player.OpenCombs)
		}
	}

	// 进入请统阶段
	sm.game.Phase = model.PhaseTongAsk
	sm.game.TongRequested = true
	sm.game.TongOrder = []int{2, 1, 0} // 闲二 -> 闲一 -> 庄家
	sm.game.TongCurrent = 0

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyTongAsk,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"order": sm.game.TongOrder,
		},
	})

	sm.broadcastPlayerHand(0)
	sm.notifyTongTurn()
}

// notifyTongTurn 通知当前统牌抉择人
func (sm *StateMachine) notifyTongTurn() {
	if sm.game.TongCurrent >= len(sm.game.TongOrder) {
		// 所有人抉择完毕, 进入打牌阶段
		sm.enterPlayPhase()
		return
	}

	playerIdx := sm.game.TongOrder[sm.game.TongCurrent]
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyTongTurn,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"player_index": playerIdx,
			"player_id":    sm.game.Players[playerIdx].ID,
		},
	})
}
