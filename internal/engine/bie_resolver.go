package engine

import "github.com/cwr0401/f3moon/internal/model"

// BieAssignment 别字分配方案
type BieAssignment struct {
	TileID int
	AsName model.TileName // 别视作的牌名(花三/花五/花七)
}

// EnumerateBieAssignments 枚举所有别字的分配方案
// 每张别字可作: 花三, 花五, 花七 (3种)
// n张别字有 3^n 种方案
func EnumerateBieAssignments(bieTiles []*model.Tile) [][]BieAssignment {
	if len(bieTiles) == 0 {
		return [][]BieAssignment{{}}
	}

	choices := []model.TileName{model.TileSan, model.TileWu, model.TileQi}
	var results [][]BieAssignment

	var backtrack func(idx int, current []BieAssignment)
	backtrack = func(idx int, current []BieAssignment) {
		if idx == len(bieTiles) {
			copied := make([]BieAssignment, len(current))
			copy(copied, current)
			results = append(results, copied)
			return
		}
		for _, choice := range choices {
			current = append(current, BieAssignment{
				TileID: bieTiles[idx].ID,
				AsName: choice,
			})
			backtrack(idx+1, current)
			current = current[:len(current)-1]
		}
	}

	backtrack(0, nil)
	return results
}

// ApplyBieAssignments 将别字分配方案应用到手牌计数器
// 返回一个扩展后的计数器, 其中别字被替换为对应的经牌
func ApplyBieAssignments(counter map[model.TileName]int, assignments []BieAssignment) map[model.TileName]int {
	result := make(map[model.TileName]int)
	for k, v := range counter {
		result[k] = v
	}
	// 减少别字数量, 增加对应的经牌数量
	for _, a := range assignments {
		result[model.TileBie]--
		if result[model.TileBie] <= 0 {
			delete(result, model.TileBie)
		}
		result[a.AsName]++
	}
	return result
}

// BuildBieMapping 从分配方案构建组合的BieMapping
func BuildBieMapping(assignments []BieAssignment) map[int]model.TileName {
	mapping := make(map[int]model.TileName)
	for _, a := range assignments {
		mapping[a.TileID] = a.AsName
	}
	return mapping
}

// FilterBieTiles 从手牌中筛选出别字
func FilterBieTiles(tiles []*model.Tile) []*model.Tile {
	var bieTiles []*model.Tile
	for _, t := range tiles {
		if t.Name == model.TileBie {
			bieTiles = append(bieTiles, t)
		}
	}
	return bieTiles
}

// CountTiles 统计各牌面名称的数量
func CountTiles(tiles []*model.Tile) map[model.TileName]int {
	counter := make(map[model.TileName]int)
	for _, t := range tiles {
		counter[t.Name]++
	}
	return counter
}

// CountTilesWithBie 统计各牌面名称的数量(别字按分配方案替换)
func CountTilesWithBie(tiles []*model.Tile, assignments []BieAssignment) map[model.TileName]int {
	bieMap := BuildBieMapping(assignments)
	counter := make(map[model.TileName]int)
	for _, t := range tiles {
		if t.Name == model.TileBie {
			if mapped, ok := bieMap[t.ID]; ok {
				counter[mapped]++
			} else {
				counter[model.TileBie]++
			}
		} else {
			counter[t.Name]++
		}
	}
	return counter
}
