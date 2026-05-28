package game

import "github.com/cwr0401/f3moon/internal/model"

// EventType 事件类型
type EventType int

const (
	EventJoin      EventType = iota // 加入房间
	EventLeave                      // 离开房间
	EventStart                      // 开始游戏
	EventCut                        // 切牌
	EventTong                       // 统牌抉择
	EventDraw                       // 起牌
	EventDiscard                    // 出牌
	EventPair                       // 碰牌(对/招/泛)
	EventGanta                      // 赶塔
	EventWin                        // 和牌
	EventPass                       // 过
	EventDangJing                   // 选择当经
)

// GameEvent 游戏事件
type GameEvent struct {
	Type     EventType    `json:"type"`
	PlayerID string       `json:"player_id"`
	Data     interface{}  `json:"data"`
}

// CutData 切牌事件数据
type CutData struct {
	Position int `json:"position"` // 切牌位置(37-110之间)
}

// DiscardData 出牌事件数据
type DiscardData struct {
	TileID int `json:"tile_id"`
}

// PairData 碰牌事件数据
type PairData struct {
	TileID    int `json:"tile_id"`     // 打出的牌ID
	PairSize  int `json:"pair_size"`   // 3=对, 4=招, 5=泛
}

// TongData 统牌事件数据
type TongData struct {
	TileName model.TileName `json:"tile_name"` // 统的牌名
	TongSize int            `json:"tong_size"`  // 4=统, 5=五张统
}

// DangJingData 当经选择数据
type DangJingData struct {
	Jing model.TileName `json:"jing"`
}

// NotifyType 推送通知类型
type NotifyType string

const (
	NotifyGameStart   NotifyType = "game:start"
	NotifyCutWait     NotifyType = "game:cut-wait"
	NotifyDealt       NotifyType = "game:dealt"
	NotifyTongAsk     NotifyType = "game:tong-ask"
	NotifyTongTurn    NotifyType = "game:tong-turn"
	NotifyTurn        NotifyType = "game:turn"
	NotifyDraw        NotifyType = "game:draw"
	NotifyDiscard     NotifyType = "game:discard"
	NotifyPair        NotifyType = "game:pair"
	NotifyGanta       NotifyType = "game:ganta"
	NotifyWin         NotifyType = "game:win"
	NotifyHuang       NotifyType = "game:huang"
	NotifyCheck       NotifyType = "game:check"
	NotifyPlayerHand  NotifyType = "player:hand"
	NotifyPhaseChange NotifyType = "game:phase"
)

// NotifyMessage 推送消息
type NotifyMessage struct {
	Type    NotifyType   `json:"type"`
	GameID  string       `json:"game_id"`
	Player  string       `json:"player"` // 目标玩家(空=全体)
	Data    interface{}  `json:"data"`
}

// NotifyBroadcaster 推送接口
type NotifyBroadcaster interface {
	Broadcast(msg NotifyMessage)
}
