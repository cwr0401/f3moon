package engine

import (
	"github.com/cwr0401/f3moon/internal/model"
)

// CalcCombHu 计算一个组合的胡数
func CalcCombHu(comb *model.Combination, dangJing model.TileName) int {
	switch comb.Type {
	case model.CombWord, model.CombNumeric:
		return calcSentenceHu(comb, dangJing)
	case model.CombSingle:
		return calcSingleHu(comb, dangJing)
	}
	return 0
}

// calcSentenceHu 一句(文字/数字组合)的胡数计算
// = 基础胡(黑0/红1) + 非经花加成 + 经牌胡数
func calcSentenceHu(comb *model.Combination, dangJing model.TileName) int {
	baseHu := 0
	if comb.HasRedTile() {
		baseHu = 1 // 红色一句1胡
	}

	flowerBonus := 0
	jingHu := 0

	for _, tile := range comb.Tiles {
		resolved := comb.ResolveName(tile)
		if model.IsJingName(resolved) {
			jingHu += calcJingTileHu(tile, resolved, dangJing)
		} else if tile.IsFlower {
			flowerBonus++
		}
	}

	return baseHu + flowerBonus + jingHu
}

// calcJingTileHu 经牌单字胡数
// 白经: 当经2胡, 不当经1胡
// 花经: 当经4胡, 不当经2胡
func calcJingTileHu(tile *model.Tile, resolved model.TileName, dangJing model.TileName) int {
	isDang := (resolved == dangJing)
	if tile.IsFlower || tile.Name == model.TileBie {
		if isDang {
			return 4
		}
		return 2
	}
	// 白经
	if isDang {
		return 2
	}
	return 1
}

// calcSingleHu 单字组合(坎/统/五张统)的胡数
func calcSingleHu(comb *model.Combination, dangJing model.TileName) int {
	size := comb.SingleSize()
	if size == 0 {
		return 0
	}

	// 判断是否为经牌组合
	firstResolved := comb.ResolveName(comb.Tiles[0])
	if model.IsJingName(firstResolved) {
		return calcJingSingleHu(comb, dangJing)
	}

	// 非经牌
	color := model.GetTileColor(comb.Tiles[0].Name)
	flowerCount := 0
	for _, t := range comb.Tiles {
		if t.IsFlower {
			flowerCount++
		}
	}

	return calcNonJingSingleHu(size, color, flowerCount)
}

// calcNonJingSingleHu 非经牌单字组合胡数
func calcNonJingSingleHu(size int, color model.TileColor, flowerCount int) int {
	var baseHu int
	switch size {
	case 3: // 坎
		if color == model.ColorRed {
			baseHu = 2
		} else {
			baseHu = 1 // 黑色坎1胡
		}
	case 4: // 统
		if color == model.ColorRed {
			baseHu = 4
		} else {
			baseHu = 2 // 黑色统2胡
		}
	case 5: // 五张统
		if color == model.ColorRed {
			baseHu = 8
		} else {
			baseHu = 4 // 黑色五张统4胡
		}
	default:
		return 0
	}

	// 红色坎不叠加花加成(红色牌无花), 黑色坎叠加
	if color == model.ColorRed {
		return baseHu
	}

	// 黑色牌的花加成: 每花+1胡
	return baseHu + flowerCount
}

// calcJingSingleHu 经牌单字组合胡数(查表)
func calcJingSingleHu(comb *model.Combination, dangJing model.TileName) int {
	size := comb.SingleSize()

	flowerCount := 0
	bieCount := 0
	for _, t := range comb.Tiles {
		if t.Name == model.TileBie {
			bieCount++
		} else if t.IsFlower {
			flowerCount++
		}
	}

	pair := LookupJingHu(size, flowerCount, bieCount)
	resolved := comb.ResolveName(comb.Tiles[0])
	if resolved == dangJing {
		return pair.Dang
	}
	return pair.BuDang
}

// CalcSingleTileHu 计算单张牌的胡数(用于不完整组合中的单字)
func CalcSingleTileHu(tile *model.Tile, resolved model.TileName, dangJing model.TileName) int {
	if model.IsJingName(resolved) {
		return calcJingTileHu(tile, resolved, dangJing)
	}
	if tile.IsFlower {
		return 1 // 带花非经牌1胡
	}
	return 0 // 不带花非经牌0胡(注: 白经1胡已在上面的calcJingTileHu中处理)
}

// CalcTotalHu 计算玩家全部胡数(手牌组合 + 明牌区)
func CalcTotalHu(handCombs []*model.Combination, openCombs []*model.Combination, dangJing model.TileName) int {
	total := 0
	for _, comb := range handCombs {
		total += CalcCombHu(comb, dangJing)
	}
	for _, comb := range openCombs {
		total += CalcOpenCombHu(comb, dangJing)
	}
	return total
}

// CalcPo 计算爬坡数
// 17-21=1坡, 22-26=2坡, 每5胡增1坡
func CalcPo(totalHu int) int {
	if totalHu < 17 {
		return 0
	}
	return (totalHu-17)/5 + 1
}

// CanWin 胡数是否达到和牌条件(>=17)
func CanWin(totalHu int) bool {
	return totalHu >= 17
}
