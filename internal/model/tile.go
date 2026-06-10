package model

import "sync"

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

// tileMap 单例牌组映射 (只读)
var tileMap map[uint8]Tile
var tileMapOnce sync.Once

// initTileMap 按 design/tile_code.md 初始化牌组映射
func initTileMap() {
	tileMap = make(map[uint8]Tile)

	// 乙 (0-4): 黑, 前3张无花, 后2张有花
	for id := 0; id <= 4; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileYi,
			Color:    ColorBlack,
			IsFlower: id >= 3,
			Numeric:  1,
		}
	}
	// 二 (5-9): 黑, 无花
	for id := 5; id <= 9; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileEr,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  2,
		}
	}
	// 三 (10-14): 红, 前3张无花, 后2张有花
	for id := 10; id <= 14; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileSan,
			Color:    ColorRed,
			IsFlower: id >= 13,
			Numeric:  3,
		}
	}
	// 四 (15-19): 黑, 无花
	for id := 15; id <= 19; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileSi,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  4,
		}
	}
	// 五 (20-24): 红, 前3张无花, 后2张有花
	for id := 20; id <= 24; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileWu,
			Color:    ColorRed,
			IsFlower: id >= 23,
			Numeric:  5,
		}
	}
	// 六 (25-29): 黑, 无花
	for id := 25; id <= 29; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileLiu,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  6,
		}
	}
	// 七 (30-34): 红, 前3张无花, 后2张有花
	for id := 30; id <= 34; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileQi,
			Color:    ColorRed,
			IsFlower: id >= 33,
			Numeric:  7,
		}
	}
	// 八 (35-39): 黑, 无花
	for id := 35; id <= 39; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileBa,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  8,
		}
	}
	// 九 (40-44): 黑, 前3张无花, 后2张有花
	for id := 40; id <= 44; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileJiu,
			Color:    ColorBlack,
			IsFlower: id >= 43,
			Numeric:  9,
		}
	}
	// 十 (45-49): 黑, 无花
	for id := 45; id <= 49; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileShi,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  10,
		}
	}
	// 上 (50-54): 红, 无花
	for id := 50; id <= 54; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileShang,
			Color:    ColorRed,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 大 (55-59): 红, 无花
	for id := 55; id <= 59; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileDa,
			Color:    ColorRed,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 人 (60-64): 红, 无花
	for id := 60; id <= 64; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileRen,
			Color:    ColorRed,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 可 (65-69): 红, 无花
	for id := 65; id <= 69; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileKe,
			Color:    ColorRed,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 知 (70-74): 红, 无花
	for id := 70; id <= 74; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileZhi,
			Color:    ColorRed,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 礼 (75-79): 红, 无花
	for id := 75; id <= 79; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileLi,
			Color:    ColorRed,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 孔 (80-84): 黑, 无花
	for id := 80; id <= 84; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileKong,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 己 (85-89): 黑, 无花
	for id := 85; id <= 89; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileJi,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 化 (90-94): 黑, 无花
	for id := 90; id <= 94; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileHua,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 千 (95-99): 黑, 无花
	for id := 95; id <= 99; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileQian,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 土 (100-104): 黑, 无花
	for id := 100; id <= 104; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileTu,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 子 (105-109): 黑, 无花
	for id := 105; id <= 109; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileZi,
			Color:    ColorBlack,
			IsFlower: false,
			Numeric:  0,
		}
	}
	// 别 (110-111): 红, 有花
	for id := 110; id <= 111; id++ {
		tileMap[uint8(id)] = Tile{
			ID:       id,
			Name:     TileBie,
			Color:    ColorRed,
			IsFlower: true,
			Numeric:  0,
		}
	}
}

// GetTile 通过 ID 获取牌 (只读访问)
func GetTile(id uint8) *Tile {
	tileMapOnce.Do(initTileMap)
	if tile, ok := tileMap[id]; ok {
		return &tile
	}
	return nil
}

// GetAllTileIDs 返回所有牌的 ID 数组 (0-111)
func GetAllTileIDs() []uint8 {
	tileMapOnce.Do(initTileMap)
	ids := make([]uint8, 0, 112)
	for id := uint8(0); id < 112; id++ {
		ids = append(ids, id)
	}
	return ids
}

// GetTilesFromIDs 通过 ID 数组获取牌对象数组
func GetTilesFromIDs(ids []uint8) []*Tile {
	tileMapOnce.Do(initTileMap)
	tiles := make([]*Tile, 0, len(ids))
	for _, id := range ids {
		if tile := GetTile(id); tile != nil {
			tiles = append(tiles, tile)
		}
	}
	return tiles
}
