package game

import (
	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// handleDraw 处理起牌
func (sm *StateMachine) handleDraw(evt GameEvent) error {
	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 || playerIdx != sm.game.CurrentTurn {
		return nil
	}

	tile := sm.game.DrawFromTop()
	if tile == nil {
		return nil
	}

	player := sm.game.Players[playerIdx]
	player.AddTileToHand(tile)

	// 检查自摸
	if engine.CheckSelfWin(player.Hand, player.OpenCombs, player.DangJing) {
		sm.game.Winner = playerIdx
		sm.game.WinType = model.WinTypeZiMo
		sm.game.Phase = model.PhaseCheck
		sm.broadcast.Broadcast(NotifyMessage{
			Type:   NotifyWin,
			GameID: sm.game.ID,
			Data: map[string]interface{}{
				"winner":   playerIdx,
				"win_type": model.WinTypeZiMo,
			},
		})
		return nil
	}

	// 检查赶塔
	if engine.CanGanta(player, tile) {
		// 等待玩家选择赶塔或出牌
		sm.broadcast.Broadcast(NotifyMessage{
			Type:   NotifyTurn,
			GameID: sm.game.ID,
			Data: map[string]interface{}{
				"current_turn": playerIdx,
				"player_id":    player.ID,
				"action":       "discard_or_ganta",
			},
		})
		return nil
	}

	// 通知起牌, 等待出牌
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyDraw,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"player_index": playerIdx,
		},
	})
	sm.broadcastPlayerHand(playerIdx)

	return nil
}

// handleDiscard 处理出牌
func (sm *StateMachine) handleDiscard(evt GameEvent) error {
	data, ok := evt.Data.(DiscardData)
	if !ok {
		return nil
	}

	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 || playerIdx != sm.game.CurrentTurn {
		return nil
	}

	player := sm.game.Players[playerIdx]
	tile := player.RemoveTileFromHand(data.TileID)
	if tile == nil {
		return nil
	}

	sm.game.AddToDiscard(tile)
	sm.game.LastDiscarder = playerIdx

	// 广播出牌
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyDiscard,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"tile":         tile,
			"player_index": playerIdx,
		},
	})

	// 海底捞月检查
	if sm.game.IsHaiDi() {
		sm.handleHaiDi()
		return nil
	}

	// 按优先级询问其他玩家
	sm.checkOtherPlayers(playerIdx)

	return nil
}

// handlePair 处理碰牌
func (sm *StateMachine) handlePair(evt GameEvent) error {
	data, ok := evt.Data.(PairData)
	if !ok {
		return nil
	}

	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 {
		return nil
	}

	player := sm.game.Players[playerIdx]
	if !player.CanPair() {
		return nil
	}

	// 从出牌区取回打出的牌
	discardTile := sm.game.LastDiscard
	if discardTile == nil {
		return nil
	}

	// 从手牌中取出与打出的牌同名的牌
	var pairTiles []*model.Tile
	pairTiles = append(pairTiles, discardTile) // 打出的牌

	needCount := data.PairSize - 1 // 需要从手牌中取的数量
	remaining := make([]*model.Tile, 0, len(player.Hand))
	collected := 0

	for _, t := range player.Hand {
		if t.Name == discardTile.Name && collected < needCount {
			pairTiles = append(pairTiles, t)
			collected++
		} else {
			remaining = append(remaining, t)
		}
	}

	if collected < needCount {
		return nil // 手牌不够
	}

	player.Hand = remaining
	player.IncrementPair()

	// 创建明牌区组合
	comb := model.NewCombination(model.CombSingle, model.CombComplete, pairTiles)
	player.OpenCombs = append(player.OpenCombs, comb)
	player.OpenTiles = append(player.OpenTiles, pairTiles...)

	// 从出牌区移除
	if len(sm.game.DiscardPile) > 0 {
		sm.game.DiscardPile = sm.game.DiscardPile[:len(sm.game.DiscardPile)-1]
	}

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyPair,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"player_index": playerIdx,
			"pair_size":    data.PairSize,
			"tile_name":    discardTile.Name,
		},
	})

	// 招/泛: 从公牌底部起1张牌, 然后出牌
	if data.PairSize >= 4 {
		drawn := sm.game.DrawFromBottom()
		if drawn != nil {
			player.AddTileToHand(drawn)
		}
	}

	sm.broadcastPlayerHand(playerIdx)

	// 对牌后需要出牌, 设置当前轮到该玩家
	sm.game.CurrentTurn = playerIdx
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyTurn,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"current_turn": playerIdx,
			"player_id":    player.ID,
			"action":       "discard",
		},
	})

	return nil
}

// handleGanta 处理赶塔
func (sm *StateMachine) handleGanta(evt GameEvent) error {
	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 {
		return nil
	}

	player := sm.game.Players[playerIdx]

	// 找到明牌区中的统(4张)
	var targetComb *model.Combination
	var targetIdx int
	for i, comb := range player.OpenCombs {
		if comb.Type == model.CombSingle && len(comb.Tiles) == 4 {
			// 检查新起的牌是否与统相同
			for _, t := range player.Hand {
				if t.Name == comb.Tiles[0].Name {
					targetComb = comb
					targetIdx = i
					// 从手牌移除该牌
					player.RemoveTileFromHand(t.ID)
					// 加入统
					targetComb.Tiles = append(targetComb.Tiles, t)
					break
				}
			}
			if targetComb != nil {
				break
			}
		}
	}

	if targetComb == nil {
		return nil
	}

	player.OpenCombs[targetIdx] = targetComb

	// 从公牌底部起1张牌
	drawn := sm.game.DrawFromBottom()
	if drawn != nil {
		player.AddTileToHand(drawn)
	}

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyGanta,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"player_index": playerIdx,
			"tile_name":    targetComb.Tiles[0].Name,
		},
	})
	sm.broadcastPlayerHand(playerIdx)

	// 赶塔后需出牌
	sm.game.CurrentTurn = playerIdx
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyTurn,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"current_turn": playerIdx,
			"player_id":    player.ID,
			"action":       "discard",
		},
	})

	return nil
}

// handleWin 处理和牌
func (sm *StateMachine) handleWin(evt GameEvent) error {
	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 {
		return nil
	}

	player := sm.game.Players[playerIdx]

	// 验证和牌
	if !engine.CheckSelfWin(player.Hand, player.OpenCombs, player.DangJing) {
		return nil
	}

	sm.game.Winner = playerIdx
	sm.game.WinType = model.WinTypeZiMo
	sm.game.Phase = model.PhaseCheck

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyWin,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"winner":   playerIdx,
			"win_type": model.WinTypeZiMo,
		},
	})

	return nil
}

// handlePass 处理过
func (sm *StateMachine) handlePass(evt GameEvent) error {
	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 {
		return nil
	}

	// 如果是出牌后的询问阶段, 检查下一个优先级的玩家
	sm.advanceCheckAfterDiscard(playerIdx)
	return nil
}

// checkOtherPlayers 出牌后按优先级询问其他玩家
func (sm *StateMachine) checkOtherPlayers(discarder int) {
	order := sm.game.PriorityOrder(discarder)
	sm.checkPlayerInOrder(order, 0)
}

// checkPlayerInOrder 按顺序检查玩家
func (sm *StateMachine) checkPlayerInOrder(order []int, idx int) {
	if idx >= len(order) {
		// 所有人都过, 下一个玩家起牌
		sm.advanceTurn()
		return
	}

	playerIdx := order[idx]
	player := sm.game.Players[playerIdx]
	discardTile := sm.game.LastDiscard

	// 检查和牌(优先级最高)
	if engine.CheckDiscardWin(player.Hand, discardTile, player.OpenCombs, player.DangJing) {
		// 通知该玩家可以和牌
		sm.broadcast.Broadcast(NotifyMessage{
			Type:   NotifyTurn,
			GameID: sm.game.ID,
			Data: map[string]interface{}{
				"current_turn": playerIdx,
				"player_id":    player.ID,
				"action":       "win_or_pass",
				"discard_tile": discardTile,
			},
		})
		return
	}

	// 检查碰牌
	canPair, pairSize := engine.CanPairWith(player.Hand, discardTile)
	if canPair && player.CanPair() {
		sm.broadcast.Broadcast(NotifyMessage{
			Type:   NotifyTurn,
			GameID: sm.game.ID,
			Data: map[string]interface{}{
				"current_turn": playerIdx,
				"player_id":    player.ID,
				"action":       "pair_or_pass",
				"pair_size":    pairSize,
				"discard_tile": discardTile,
			},
		})
		return
	}

	// 下一个玩家
	sm.checkPlayerInOrder(order, idx+1)
}

// advanceCheckAfterDiscard 过牌后继续检查
func (sm *StateMachine) advanceCheckAfterDiscard(passedPlayerIdx int) {
	discarder := sm.game.LastDiscarder
	order := sm.game.PriorityOrder(discarder)

	// 找到当前玩家在顺序中的位置, 检查下一个
	for i, idx := range order {
		if idx == passedPlayerIdx {
			sm.checkPlayerInOrder(order, i+1)
			return
		}
	}

	// 不在顺序中, 直接推进
	sm.advanceTurn()
}

// advanceTurn 推进到下一个玩家
func (sm *StateMachine) advanceTurn() {
	nextPlayer := sm.game.NextPlayer(sm.game.CurrentTurn)
	sm.game.CurrentTurn = nextPlayer

	// 检查公牌是否耗尽
	if sm.game.DrawPileSize() == 0 {
		sm.handleHuangZhuang()
		return
	}

	// 通知下一个玩家起牌
	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyTurn,
		GameID: sm.game.ID,
		Data: map[string]interface{}{
			"current_turn": nextPlayer,
			"player_id":    sm.game.Players[nextPlayer].ID,
			"action":       "draw",
		},
	})
}

// handleHaiDi 处理海底捞月
func (sm *StateMachine) handleHaiDi() {
	// 停止碰牌、统牌、赶塔
	// 剩余3张牌按顺序分发给3个玩家
	for i := 0; i < 3; i++ {
		playerIdx := (sm.game.LastDiscarder + 1 + i) % 3
		tile := sm.game.DrawFromTop()
		if tile != nil {
			player := sm.game.Players[playerIdx]
			player.AddTileToHand(tile)

			// 检查和牌
			if engine.CheckSelfWin(player.Hand, player.OpenCombs, player.DangJing) {
				sm.game.Winner = playerIdx
				sm.game.WinType = model.WinTypeHaiDi
				sm.game.Phase = model.PhaseCheck

				sm.broadcast.Broadcast(NotifyMessage{
					Type:   NotifyWin,
					GameID: sm.game.ID,
					Data: map[string]interface{}{
						"winner":   playerIdx,
						"win_type": model.WinTypeHaiDi,
					},
				})
				return
			}
		}
	}

	// 无人和牌, 黄庄
	sm.handleHuangZhuang()
}

// handleHuangZhuang 处理黄庄
func (sm *StateMachine) handleHuangZhuang() {
	sm.game.Phase = model.PhaseFinished
	sm.game.Winner = -1

	sm.broadcast.Broadcast(NotifyMessage{
		Type:   NotifyHuang,
		GameID: sm.game.ID,
		Data:   map[string]interface{}{},
	})
}

// handleDangJing 处理当经选择
func (sm *StateMachine) handleDangJing(evt GameEvent) error {
	data, ok := evt.Data.(DangJingData)
	if !ok {
		return nil
	}

	playerIdx := sm.game.PlayerIndexByID(evt.PlayerID)
	if playerIdx < 0 {
		return nil
	}

	sm.game.Players[playerIdx].DangJing = data.Jing
	sm.broadcastPlayerHand(playerIdx)
	return nil
}
