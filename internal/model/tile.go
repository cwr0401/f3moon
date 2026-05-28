package model

// TileName 牌面名称
type TileName string

const (
	// 数字牌 1-10
	TileYi    TileName = "乙" // 1
	TileEr    TileName = "二" // 2
	TileSan   TileName = "三" // 3
	TileSi    TileName = "四" // 4
	TileWu    TileName = "五" // 5
	TileLiu   TileName = "六" // 6
	TileQi    TileName = "七" // 7
	TileBa    TileName = "八" // 8
	TileJiu   TileName = "九" // 9
	TileShi   TileName = "十" // 10
	// 文字牌
	TileKong  TileName = "孔"
	TileJi    TileName = "己"
	TileHua   TileName = "化"
	TileQian  TileName = "千"
	TileTu    TileName = "土"
	TileZi    TileName = "子"
	TileShang TileName = "上"
	TileDa    TileName = "大"
	TileRen   TileName = "人"
	TileKe    TileName = "可"
	TileZhi   TileName = "知"
	TileLi    TileName = "礼"
	// 特殊牌
	TileBie   TileName = "别"
)

// TileColor 牌色
type TileColor int

const (
	ColorBlack TileColor = iota
	ColorRed
)

// Tile 一张牌的完整属性
type Tile struct {
	ID       int      `json:"id"`
	Name     TileName `json:"name"`
	Color    TileColor `json:"color"`
	IsFlower bool     `json:"is_flower"`
	Numeric  int      `json:"numeric"` // 数字值(仅数字牌有效, 1-10; 文字牌/别=0)
}

// IsJing 是否为经牌(三、五、七)
func (t *Tile) IsJing() bool {
	return t.Name == TileSan || t.Name == TileWu || t.Name == TileQi
}

// AllTileNames 所有牌面名称(不含别)
var AllTileNames = []TileName{
	TileYi, TileEr, TileSan, TileSi, TileWu,
	TileLiu, TileQi, TileBa, TileJiu, TileShi,
	TileKong, TileJi, TileHua, TileQian, TileTu, TileZi,
	TileShang, TileDa, TileRen, TileKe, TileZhi, TileLi,
}

// NumericTiles 数字牌名称
var NumericTiles = []TileName{
	TileYi, TileEr, TileSan, TileSi, TileWu,
	TileLiu, TileQi, TileBa, TileJiu, TileShi,
}

// WordTiles 文字牌名称
var WordTiles = []TileName{
	TileKong, TileJi, TileHua, TileQian, TileTu, TileZi,
	TileShang, TileDa, TileRen, TileKe, TileZhi, TileLi,
}

// JingTiles 经牌名称
var JingTiles = []TileName{TileSan, TileWu, TileQi}

// BlackTileNames 黑色牌名称
var BlackTileNames = map[TileName]bool{
	TileYi: true, TileEr: true, TileSi: true, TileLiu: true,
	TileBa: true, TileJiu: true, TileShi: true,
	TileKong: true, TileJi: true, TileHua: true,
	TileQian: true, TileTu: true, TileZi: true,
}

// RedTileNames 红色牌名称
var RedTileNames = map[TileName]bool{
	TileSan: true, TileWu: true, TileQi: true,
	TileShang: true, TileDa: true, TileRen: true,
	TileKe: true, TileZhi: true, TileLi: true,
}

// FlowerTileNames 带花的牌名称(5张中有2张带花)
var FlowerTileNames = map[TileName]bool{
	TileYi: true, TileSan: true, TileWu: true, TileQi: true, TileJiu: true,
}

// NumericValue 牌面名称对应的数字值
func NumericValue(name TileName) int {
	switch name {
	case TileYi:
		return 1
	case TileEr:
		return 2
	case TileSan:
		return 3
	case TileSi:
		return 4
	case TileWu:
		return 5
	case TileLiu:
		return 6
	case TileQi:
		return 7
	case TileBa:
		return 8
	case TileJiu:
		return 9
	case TileShi:
		return 10
	default:
		return 0
	}
}

// GetTileColor 获取牌面颜色
func GetTileColor(name TileName) TileColor {
	if RedTileNames[name] || name == TileBie {
		return ColorRed
	}
	return ColorBlack
}

// IsNumericTile 是否为数字牌
func IsNumericTile(name TileName) bool {
	return NumericValue(name) > 0
}

// IsJingName 是否为经牌名称
func IsJingName(name TileName) bool {
	return name == TileSan || name == TileWu || name == TileQi
}
