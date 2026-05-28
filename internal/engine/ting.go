package engine

import (
	"github.com/cwr0401/f3moon/internal/model"
)

// TingResult 听牌检测结果
type TingResult struct {
	Arrangement *Arrangement   // 听牌时的牌型排列
	TingTiles   []model.TileName // 听字列表
	TingType    TingType       // 听牌牌型
	TotalHu     int            // 当前总胡数
}

// DetectTing 检测手牌是否听牌
func DetectTing(hand []*model.Tile, openCombs []*model.Combination, dangJing model.TileName) []*TingResult {
	arrangements := ArrangeTiles(hand, dangJing)
	var results []*TingResult

	for _, arr := range arrangements {
		tingType := arr.GetTingType()
		if tingType == TingTypeNone {
			continue
		}

		tingTiles := calcTingTiles(arr, dangJing)
		if len(tingTiles) < 2 {
			continue
		}

		totalHu := arr.TotalHu
		for _, comb := range openCombs {
			totalHu += CalcOpenCombHu(comb, dangJing)
		}

		results = append(results, &TingResult{
			Arrangement: arr,
			TingTiles:   tingTiles,
			TingType:    tingType,
			TotalHu:     totalHu,
		})
	}

	return results
}

// calcTingTiles 计算听字
func calcTingTiles(arr *Arrangement, dangJing model.TileName) []model.TileName {
	tingSet := make(map[model.TileName]bool)

	switch arr.GetTingType() {
	case TingTypeOne:
		// 2个不完整组合, 每个不完整组合找出能补全的牌
		for _, inc := range arr.Incomplete {
			completions := findCompletions(inc, dangJing)
			for _, t := range completions {
				tingSet[t] = true
			}
		}
	case TingTypeTwo:
		// 1个单字, 找出能与单字组成不完整组合的牌
		if len(arr.Singles) > 0 {
			completions := findCompletionsFromSingle(arr.Singles[0], dangJing)
			for _, t := range completions {
				tingSet[t] = true
			}
		}
	}

	// 如果听字含三/五/七, 别也是听字
	for name := range tingSet {
		if model.IsJingName(name) {
			tingSet[model.TileBie] = true
		}
	}

	var tingTiles []model.TileName
	for name := range tingSet {
		tingTiles = append(tingTiles, name)
	}
	return tingTiles
}

// findCompletions 找出能补全不完整组合的牌
func findCompletions(inc *model.Combination, dangJing model.TileName) []model.TileName {
	var completions []model.TileName
	completionSet := make(map[model.TileName]bool)

	switch inc.Type {
	case model.CombWord:
		// 找出文字组合中缺失的那张牌
		for _, word := range model.WordCombinations {
			if isSubArrayOfWord(inc, word) {
				missing := findMissingFromWord(inc, word)
				for _, m := range missing {
					completionSet[m] = true
				}
			}
		}
	case model.CombNumeric:
		// 找出数字组合中缺失的那张牌
		for _, nums := range model.NumericCombinations {
			if isSubArrayOfNumeric(inc, nums) {
				missing := findMissingFromNumeric(inc, nums)
				for _, m := range missing {
					completionSet[m] = true
				}
			}
		}
	case model.CombSingle:
		// 对(不完整坎), 缺少第3张
		if len(inc.Tiles) >= 1 {
			name := inc.ResolveName(inc.Tiles[0])
			completionSet[name] = true
		}
	}

	for name := range completionSet {
		completions = append(completions, name)
	}
	return completions
}

// findCompletionsFromSingle 找出能与单字组成不完整组合的牌
func findCompletionsFromSingle(single *model.Tile, dangJing model.TileName) []model.TileName {
	var completions []model.TileName
	completionSet := make(map[model.TileName]bool)
	name := single.Name

	// 与其他牌组成不完整文字组合
	for _, word := range model.WordCombinations {
		for i, wn := range word {
			if wn == name {
				// 该字出现在这个文字组合中, 其他2张都可以作为听字
				for j, other := range word {
					if j != i {
						completionSet[other] = true
					}
				}
			}
		}
	}

	// 与其他牌组成不完整数字组合
	if model.IsNumericTile(name) {
		val := model.NumericValue(name)
		for _, nums := range model.NumericCombinations {
			for i, v := range nums {
				if v == val {
					for j, other := range nums {
						if j != i {
							otherName := model.NumericNameFromValue(other)
							completionSet[otherName] = true
						}
					}
				}
			}
		}
	}

	// 与另一张相同的牌组成对(不完整坎)
	completionSet[name] = true

	// 别字与该牌组成不完整组合
	if model.IsJingName(name) || model.IsNumericTile(name) {
		completionSet[model.TileBie] = true
	}

	for n := range completionSet {
		completions = append(completions, n)
	}
	return completions
}

// isSubArrayOfWord 判断不完整组合是否为某个文字组合的子集
func isSubArrayOfWord(inc *model.Combination, word [3]model.TileName) bool {
	if len(inc.Tiles) != 2 {
		return false
	}
	tileNames := make(map[model.TileName]bool)
	for _, t := range inc.Tiles {
		tileNames[inc.ResolveName(t)] = true
	}
	wordSet := make(map[model.TileName]bool)
	for _, wn := range word {
		wordSet[wn] = true
	}
	// inc的所有牌名都在word中
	for name := range tileNames {
		if !wordSet[name] {
			return false
		}
	}
	// inc的牌数量不超过word中的数量
	incCount := make(map[model.TileName]int)
	wordCount := make(map[model.TileName]int)
	for _, t := range inc.Tiles {
		incCount[inc.ResolveName(t)]++
	}
	for _, wn := range word {
		wordCount[wn]++
	}
	for name, c := range incCount {
		if wordCount[name] < c {
			return false
		}
	}
	return true
}

// isSubArrayOfNumeric 判断不完整组合是否为某个数字组合的子集
func isSubArrayOfNumeric(inc *model.Combination, nums [3]int) bool {
	if len(inc.Tiles) != 2 {
		return false
	}
	tileValues := make(map[int]bool)
	for _, t := range inc.Tiles {
		resolved := inc.ResolveName(t)
		v := model.NumericValue(resolved)
		if v > 0 {
			tileValues[v] = true
		}
	}
	numsSet := make(map[int]bool)
	for _, v := range nums {
		numsSet[v] = true
	}
	for v := range tileValues {
		if !numsSet[v] {
			return false
		}
	}
	return true
}

// findMissingFromWord 找出文字组合中缺失的牌
func findMissingFromWord(inc *model.Combination, word [3]model.TileName) []model.TileName {
	incNames := make(map[model.TileName]int)
	for _, t := range inc.Tiles {
		incNames[inc.ResolveName(t)]++
	}
	wordCount := make(map[model.TileName]int)
	for _, wn := range word {
		wordCount[wn]++
	}
	var missing []model.TileName
	for wn, wc := range wordCount {
		if incNames[wn] < wc {
			missing = append(missing, wn)
		}
	}
	return missing
}

// findMissingFromNumeric 找出数字组合中缺失的牌
func findMissingFromNumeric(inc *model.Combination, nums [3]int) []model.TileName {
	incValues := make(map[int]int)
	for _, t := range inc.Tiles {
		resolved := inc.ResolveName(t)
		v := model.NumericValue(resolved)
		if v > 0 {
			incValues[v]++
		}
	}
	numsCount := make(map[int]int)
	for _, v := range nums {
		numsCount[v]++
	}
	var missing []model.TileName
	for v, c := range numsCount {
		if incValues[v] < c {
			missing = append(missing, model.NumericNameFromValue(v))
		}
	}
	return missing
}
