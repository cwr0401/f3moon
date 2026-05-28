package model

// GameMode 游戏模式
type GameMode int

const (
	GameMode4Player GameMode = iota // 4人模式(含歇家)
	GameMode3Player                 // 3人定庄模式
)

// GamePhase 游戏阶段
type GamePhase int

const (
	PhaseWaiting  GamePhase = iota // 等待玩家就位
	PhaseShuffle                   // 洗牌
	PhaseCut                       // 切牌(腰牌)
	PhaseDeal                      // 起牌
	PhaseTongAsk                   // 请统
	PhasePlay                      // 打牌
	PhaseCheck                     // 和牌查验
	PhaseFinished                  // 牌局结束
)

// WinType 和牌方式
type WinType int

const (
	WinTypeZiMo      WinType = iota // 自摸
	WinTypeDianPao                  // 点炮
	WinTypeTianHu                   // 天胡
	WinTypeHaiDi                    // 海底捞月
)

// TurnAction 玩家操作
type TurnAction int

const (
	ActionDraw    TurnAction = iota // 起牌
	ActionDiscard                   // 出牌
	ActionPair                      // 对
	ActionZhao                      // 招
	ActionFan                       // 泛
	ActionGanta                     // 赶塔
	ActionTong                      // 统牌
	ActionWin                       // 和牌
	ActionPass                      // 过
	ActionDangJing                  // 选择当经
)

// GameState 牌局状态
type GameState struct {
	ID           string     `json:"id"`
	Phase        GamePhase  `json:"phase"`
	Mode         GameMode   `json:"mode"`
	Players      [4]*Player `json:"players"`       // 0=庄, 1=闲一, 2=闲二, 3=歇家(3人模式为nil)
	DrawPile     []*Tile    `json:"draw_pile"`     // 公牌
	DiscardPile  []*Tile    `json:"discard_pile"`  // 出牌区
	CurrentTurn  int        `json:"current_turn"`  // 当前出牌玩家索引
	DealerIndex  int        `json:"dealer_index"`  // 庄家索引

	TongRequested bool  `json:"tong_requested"` // 是否已请统
	TongOrder     []int `json:"tong_order"`     // 统牌抉择顺序
	TongCurrent   int   `json:"tong_current"`   // 当前统牌抉择人索引

	LastDiscard   *Tile `json:"last_discard"`   // 最后打出的牌
	LastDiscarder int   `json:"last_discarder"` // 最后出牌者索引

	Winner  int     `json:"winner"`  // 赢家索引(-1表示未定)
	WinType WinType `json:"win_type"`
}

// NewGameState 创建牌局状态
func NewGameState(id string, mode GameMode) *GameState {
	return &GameState{
		ID:          id,
		Phase:       PhaseWaiting,
		Mode:        mode,
		DrawPile:    make([]*Tile, 0),
		DiscardPile: make([]*Tile, 0),
		Winner:      -1,
	}
}

// ActivePlayers 返回参与打牌的玩家(庄,闲一,闲二)
func (g *GameState) ActivePlayers() []*Player {
	result := make([]*Player, 0, 3)
	for i := 0; i < 3; i++ {
		if g.Players[i] != nil {
			result = append(result, g.Players[i])
		}
	}
	return result
}

// NextPlayer 下一个出牌玩家索引(逆时针: 庄->闲一->闲二->庄)
func (g *GameState) NextPlayer(current int) int {
	return (current + 1) % 3
}

// DrawPileSize 公牌剩余数量
func (g *GameState) DrawPileSize() int {
	return len(g.DrawPile)
}

// DrawFromTop 从公牌顶部起一张牌
func (g *GameState) DrawFromTop() *Tile {
	if len(g.DrawPile) == 0 {
		return nil
	}
	tile := g.DrawPile[0]
	g.DrawPile = g.DrawPile[1:]
	return tile
}

// DrawFromBottom 从公牌底部起一张牌
func (g *GameState) DrawFromBottom() *Tile {
	if len(g.DrawPile) == 0 {
		return nil
	}
	idx := len(g.DrawPile) - 1
	tile := g.DrawPile[idx]
	g.DrawPile = g.DrawPile[:idx]
	return tile
}

// AddToDiscard 将牌放入出牌区
func (g *GameState) AddToDiscard(tile *Tile) {
	g.DiscardPile = append(g.DiscardPile, tile)
	g.LastDiscard = tile
}

// IsHaiDi 是否进入海底捞月阶段(公牌仅剩3张)
func (g *GameState) IsHaiDi() bool {
	return len(g.DrawPile) <= 3
}

// PlayerIndexByID 根据玩家ID查找索引
func (g *GameState) PlayerIndexByID(playerID string) int {
	for i, p := range g.Players {
		if p != nil && p.ID == playerID {
			return i
		}
	}
	return -1
}

// CutPlayer 切牌玩家索引(歇家或闲二)
func (g *GameState) CutPlayer() int {
	if g.Mode == GameMode3Player {
		return 2 // 闲二
	}
	return 3 // 歇家
}

// PriorityOrder 出牌后判断优先级的顺序
// 庄家出牌 -> 闲一, 闲二
// 闲一出牌 -> 闲二, 庄家
// 闲二出牌 -> 庄家, 闲一
func (g *GameState) PriorityOrder(discarder int) []int {
	order := make([]int, 0, 2)
	for i := 1; i <= 2; i++ {
		idx := (discarder + i) % 3
		order = append(order, idx)
	}
	return order
}
