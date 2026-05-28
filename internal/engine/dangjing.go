package engine

import (
	"github.com/cwr0401/f3moon/internal/model"
)

// ChooseDangJing 选择最优当经
// 枚举三/五/七, 计算每种选择下的总胡数, 取最大
func ChooseDangJing(hand []*model.Tile, openCombs []*model.Combination) model.TileName {
	bestJing := model.TileSan
	bestHu := -1

	for _, jing := range model.JingTiles {
		arrangement := BestArrangement(hand, jing)
		openHu := 0
		for _, comb := range openCombs {
			openHu += CalcOpenCombHu(comb, jing)
		}
		totalHu := arrangement.TotalHu + openHu
		if totalHu > bestHu {
			bestHu = totalHu
			bestJing = jing
		}
	}

	return bestJing
}
