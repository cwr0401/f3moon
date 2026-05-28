package game

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// processTongDecision 处理统牌抉择
func (sm *StateMachine) processTongDecision(evt GameEvent) error {
	if sm.game.Phase != model.PhaseTongAsk {
		return nil
	}

	currentIdx := sm.game.TongOrder[sm.game.TongCurrent]
	player := sm.game.Players[currentIdx]
	if player.ID != evt.PlayerID {
		return nil
	}

	// 如果玩家选择统牌
	if evt.Type == EventTong {
		data, ok := evt.Data.(TongData)
		if ok && data.TileName != "" {
			sm.executeTong(player, data.TileName, data.TongSize)
		}
	}

	// 下一个人抉择
	sm.game.TongCurrent++
	sm.notifyTongTurn()

	return nil
}

// executeTong 执行统牌
func (sm *StateMachine) executeTong(player *model.Player, tileName model.TileName, tongSize int) {
	// 从手牌中取出tongSize张同名牌
	var tongTiles []*model.Tile
	remaining := make([]*model.Tile, 0, len(player.Hand))

	count := 0
	for _, t := range player.Hand {
		if t.Name == tileName && count < tongSize {
			tongTiles = append(tongTiles, t)
			count++
		} else {
			remaining = append(remaining, t)
		}
	}

	if len(tongTiles) < tongSize {
		return // 不够统
	}

	player.Hand = remaining

	// 创建统组合放入明牌区
	comb := model.NewCombination(model.CombSingle, model.CombComplete, tongTiles)
	player.OpenCombs = append(player.OpenCombs, comb)
	player.OpenTiles = append(player.OpenTiles, tongTiles...)

	// 从公牌底部起牌
	var drawnTiles []*model.Tile
	switch tongSize {
	case 4: // 统: 从底部起1张
		t := sm.game.DrawFromBottom()
		if t != nil {
			drawnTiles = append(drawnTiles, t)
		}
	case 5: // 五张统: 从底部起2张
		for i := 0; i < 2; i++ {
			t := sm.game.DrawFromBottom()
			if t != nil {
				drawnTiles = append(drawnTiles, t)
			}
		}
	}

	// 起到的牌加入手牌
	for _, t := range drawnTiles {
		player.AddTileToHand(t)
	}
}

// enterPlayPhase 进入打牌阶段
func (sm *StateMachine) enterPlayPhase() {
	sm.game.Phase = model.PhasePlay
	sm.game.CurrentTurn = 0 // 庄家先出

	// 检查天胡
	player := sm.game.Players[0]
	if engine.CheckTianHu(player.Hand, player.OpenCombs, player.DangJing) {
		sm.game.Winner = 0
		sm.game.WinType = model.WinTypeTianHu
		sm.game.Phase = model.PhaseCheck
		sm.broadcast.Broadcast(NotifyMessage{
			Type:   NotifyWin,
			GameID: sm.game.ID,
			Data: map[string]interface{}{
				"winner":   0,
				"win_type": model.WinTypeTianHu,
			},
		})
		return
	}

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyTurn,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"current_turn": 0,
			"player_id":    sm.game.Players[0].ID,
			"action":       "discard",
		},
	})
}

// handlePlayTong 打牌阶段统牌
func (sm *StateMachine) handlePlayTong(evt GameEvent) error {
	data, ok := evt.Data.(TongData)
	if !ok {
		return nil
	}

	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 {
		return nil
	}

	player := sm.game.Players[playerIdx]
	sm.executeTong(player, data.TileName, data.TongSize)
	sm.broadcastPlayerHand(playerIdx)

	return nil
}
