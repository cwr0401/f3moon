package model

// PlayerRole 玩家角色
type PlayerRole int

const (
	RoleDealer PlayerRole = iota // 庄家
	RoleIdle1                    // 闲一(庄家右手边)
	RoleIdle2                    // 闲二(庄家左手边)
	RoleRest                     // 歇家
)

// Player 玩家
type Player struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Role      PlayerRole    `json:"role"`
	IsAI      bool          `json:"is_ai"`
	Hand      []*Tile       `json:"hand"`        // 手牌(仅自己可见)
	OpenTiles []*Tile       `json:"open_tiles"`   // 明牌区(公开)
	OpenCombs []*Combination `json:"open_combs"` // 明牌区组合(公开)
	DangJing  TileName      `json:"dang_jing"`    // 当经(三/五/七)
	PairCount int           `json:"pair_count"`   // 已对牌次数(上限2)
	Online    bool          `json:"online"`       // 是否在线
}

// NewPlayer 创建玩家
func NewPlayer(id, name string, role PlayerRole, isAI bool) *Player {
	return &Player{
		ID:        id,
		Name:      name,
		Role:      role,
		IsAI:      isAI,
		Hand:      make([]*Tile, 0),
		OpenTiles: make([]*Tile, 0),
		OpenCombs: make([]*Combination, 0),
		DangJing:  "",
		PairCount: 0,
		Online:    !isAI,
	}
}

// AddTileToHand 向手牌添加一张牌
func (p *Player) AddTileToHand(tile *Tile) {
	p.Hand = append(p.Hand, tile)
}

// RemoveTileFromHand 从手牌移除一张牌
func (p *Player) RemoveTileFromHand(tileID int) *Tile {
	for i, t := range p.Hand {
		if t.ID == tileID {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return t
		}
	}
	return nil
}

// FindTileInHand 在手牌中查找指定ID的牌
func (p *Player) FindTileInHand(tileID int) *Tile {
	for _, t := range p.Hand {
		if t.ID == tileID {
			return t
		}
	}
	return nil
}

// HasTileInHand 手牌中是否有指定名称的牌
func (p *Player) HasTileInHand(name TileName) int {
	count := 0
	for _, t := range p.Hand {
		if t.Name == name {
			count++
		}
	}
	return count
}

// CanPair 是否可以碰牌(还有碰牌次数)
func (p *Player) CanPair() bool {
	return p.PairCount < 2
}

// IncrementPair 增加碰牌次数
func (p *Player) IncrementPair() {
	p.PairCount++
}

// HandSize 手牌数量
func (p *Player) HandSize() int {
	return len(p.Hand)
}
