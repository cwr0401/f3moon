package model

// CombType 组合类型
type CombType int

const (
	CombWord    CombType = iota // 文字组合(3张一句)
	CombNumeric                 // 数字组合(3张连续一句)
	CombSingle                  // 单字组合(坎/统/五张统)
)

// CombCompleteness 组合完整度
type CombCompleteness int

const (
	CombComplete   CombCompleteness = iota // 完整组合
	CombIncomplete                         // 不完整组合(2张)
)

// Combination 一个牌组合
type Combination struct {
	Type         CombType         `json:"type"`
	Completeness CombCompleteness `json:"completeness"`
	Tiles        []*Tile          `json:"tiles"`
	BieMapping   map[int]TileName `json:"bie_mapping"` // tileID -> 别字视作的牌名
}

// NewCombination 创建组合
func NewCombination(combType CombType, completeness CombCompleteness, tiles []*Tile) *Combination {
	return &Combination{
		Type:         combType,
		Completeness: completeness,
		Tiles:        tiles,
		BieMapping:   make(map[int]TileName),
	}
}

// SingleSize 单字组合的牌数(3=坎, 4=统, 5=五张统)
func (c *Combination) SingleSize() int {
	if c.Type != CombSingle {
		return 0
	}
	return len(c.Tiles)
}

// IsKan 是否为坎(3张相同)
func (c *Combination) IsKan() bool {
	return c.Type == CombSingle && len(c.Tiles) == 3
}

// IsTong 是否为统(4张相同)
func (c *Combination) IsTong() bool {
	return c.Type == CombSingle && len(c.Tiles) == 4
}

// IsWuZhangTong 是否为五张统(5张相同)
func (c *Combination) IsWuZhangTong() bool {
	return c.Type == CombSingle && len(c.Tiles) == 5
}

// ResolveName 解析一张牌的实际名称(考虑别字映射)
func (c *Combination) ResolveName(tile *Tile) TileName {
	if tile.Name == TileBie {
		if mapped, ok := c.BieMapping[tile.ID]; ok {
			return mapped
		}
	}
	return tile.Name
}

// HasRedTile 组合中是否含红色牌
func (c *Combination) HasRedTile() bool {
	for _, t := range c.Tiles {
		resolved := c.ResolveName(t)
		if RedTileNames[resolved] {
			return true
		}
	}
	return false
}

// FlowerCount 组合中带花牌数(非经牌)
func (c *Combination) FlowerCount() int {
	count := 0
	for _, t := range c.Tiles {
		resolved := c.ResolveName(t)
		if t.IsFlower && !IsJingName(resolved) {
			count++
		}
	}
	return count
}

// JingFlowerCount 经牌中带花牌数
func (c *Combination) JingFlowerCount() int {
	count := 0
	for _, t := range c.Tiles {
		resolved := c.ResolveName(t)
		if (t.IsFlower || t.Name == TileBie) && IsJingName(resolved) {
			count++
		}
	}
	return count
}

// JingBieCount 经牌中别字数量
func (c *Combination) JingBieCount() int {
	count := 0
	for _, t := range c.Tiles {
		if t.Name == TileBie {
			resolved := c.ResolveName(t)
			if IsJingName(resolved) {
				count++
			}
		}
	}
	return count
}

// JingBaiCount 经牌中白经数量
func (c *Combination) JingBaiCount() int {
	count := 0
	for _, t := range c.Tiles {
		resolved := c.ResolveName(t)
		if IsJingName(resolved) && !t.IsFlower && t.Name != TileBie {
			count++
		}
	}
	return count
}

// 文字组合定义
var WordCombinations = [][3]TileName{
	{TileKong, TileYi, TileJi},   // 孔乙己
	{TileHua, TileSan, TileQian}, // 化三千
	{TileQi, TileShi, TileTu},    // 七十土
	{TileBa, TileJiu, TileZi},    // 八九子
	{TileShang, TileDa, TileRen}, // 上大人
	{TileKe, TileZhi, TileLi},    // 可知礼
}

// WordCombBlack 文字组合是否为黑色
var WordCombBlack = map[int]bool{
	0: true, // 孔乙己
	3: true, // 八九子
}

// 数字组合定义: 连续3个数字 [start, start+1, start+2]
var NumericCombinations = [][3]int{
	{1, 2, 3},   // 乙二三
	{2, 3, 4},   // 二三四
	{3, 4, 5},   // 三四五
	{4, 5, 6},   // 四五六
	{5, 6, 7},   // 五六七
	{6, 7, 8},   // 六七八
	{7, 8, 9},   // 七八九
	{8, 9, 10},  // 八九十
}

// NumericNameFromValue 数字值转牌名
func NumericNameFromValue(v int) TileName {
	switch v {
	case 1:
		return TileYi
	case 2:
		return TileEr
	case 3:
		return TileSan
	case 4:
		return TileSi
	case 5:
		return TileWu
	case 6:
		return TileLiu
	case 7:
		return TileQi
	case 8:
		return TileBa
	case 9:
		return TileJiu
	case 10:
		return TileShi
	default:
		return ""
	}
}
