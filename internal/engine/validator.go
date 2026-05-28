package engine

import (
	"github.com/cwr0401/f3moon/internal/model"
)

// ValidateWin 验证和牌是否合法
// 1. 检查牌型是否为: 1个不完整组合 + 其余完整组合
// 2. 检查胡数 >= 17
func ValidateWin(handCombs []*model.Combination, openCombs []*model.Combination, dangJing model.TileName) bool {
	incompleteCount := 0
	singleCount := 0

	for _, comb := range handCombs {
		if comb.Completeness == model.CombIncomplete {
			incompleteCount++
		}
	}

	// 和牌牌型: 恰好1个不完整组合, 无单字
	if incompleteCount != 1 || singleCount != 0 {
		return false
	}

	totalHu := CalcTotalHu(handCombs, openCombs, dangJing)
	return CanWin(totalHu)
}

// CheckSelfWin 检查玩家自己起牌后是否自摸和牌
func CheckSelfWin(hand []*model.Tile, openCombs []*model.Combination, dangJing model.TileName) bool {
	arrangements := ArrangeTiles(hand, dangJing)

	for _, arr := range arrangements {
		// 检查是否恰好1个不完整组合, 其余完整, 无单字
		if len(arr.Incomplete) == 1 && len(arr.Singles) == 0 {
			totalHu := arr.TotalHu
			for _, comb := range openCombs {
				totalHu += CalcOpenCombHu(comb, dangJing)
			}
			if CanWin(totalHu) {
				return true
			}
		}
	}
	return false
}

// CheckDiscardWin 检查其他玩家打出的牌是否能让自己和牌
func CheckDiscardWin(hand []*model.Tile, discardTile *model.Tile, openCombs []*model.Combination, dangJing model.TileName) bool {
	// 将打出的牌加入手牌后检查
	testHand := append(hand, discardTile)
	return CheckSelfWin(testHand, openCombs, dangJing)
}

// CheckTianHu 检查庄家天胡(起牌后直接和牌)
func CheckTianHu(hand []*model.Tile, openCombs []*model.Combination, dangJing model.TileName) bool {
	return CheckSelfWin(hand, openCombs, dangJing)
}

// CanPairWith 检查是否可以用打出的牌碰牌
// 返回: (能否碰, 碰牌类型: 对/招/泛)
func CanPairWith(hand []*model.Tile, discardTile *model.Tile) (bool, int) {
	if discardTile.Name == model.TileBie {
		return false, 0 // 别字不能碰
	}

	count := 0
	for _, t := range hand {
		if t.Name == discardTile.Name {
			count++
		}
	}

	switch count {
	case 2:
		return true, 3 // 对: 手中2张+打出1张=3张坎
	case 3:
		return true, 4 // 招: 手中3张+打出1张=4张统
	case 4:
		return true, 5 // 泛: 手中4张+打出1张=5张统
	}
	return false, 0
}

// CanGanta 检查是否可以赶塔
// 新起的牌与明牌区的1统(4张)相同
func CanGanta(player *model.Player, drawnTile *model.Tile) bool {
	if drawnTile.Name == model.TileBie {
		return false
	}

	for _, comb := range player.OpenCombs {
		if comb.Type == model.CombSingle && len(comb.Tiles) == 4 {
			if comb.Tiles[0].Name == drawnTile.Name {
				return true
			}
		}
	}
	return false
}

// CanTongWithHand 检查手牌中是否有可以统的牌(4张或5张相同)
func CanTongWithHand(hand []*model.Tile) (bool, []model.TileName) {
	counter := CountTiles(hand)
	var tongNames []model.TileName

	for name, count := range counter {
		if count >= 4 {
			tongNames = append(tongNames, name)
		}
	}

	return len(tongNames) > 0, tongNames
}
