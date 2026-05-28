package engine

import (
	"fmt"
	"sort"

	"github.com/cwr0401/f3moon/internal/model"
)

// Arrangement 一组牌的一种排列方案
type Arrangement struct {
	Complete   []*model.Combination // 完整组合
	Incomplete []*model.Combination // 不完整组合
	Singles    []*model.Tile        // 单字(不属于任何组合)
	TotalHu    int                  // 总胡数
}

// NewArrangement 创建排列方案
func NewArrangement() *Arrangement {
	return &Arrangement{
		Complete:   make([]*model.Combination, 0),
		Incomplete: make([]*model.Combination, 0),
		Singles:    make([]*model.Tile, 0),
	}
}

// Clone 克隆排列方案
func (a *Arrangement) Clone() *Arrangement {
	clone := &Arrangement{
		Complete:   make([]*model.Combination, len(a.Complete)),
		Incomplete: make([]*model.Combination, len(a.Incomplete)),
		Singles:    make([]*model.Tile, len(a.Singles)),
		TotalHu:    a.TotalHu,
	}
	copy(clone.Complete, a.Complete)
	copy(clone.Incomplete, a.Incomplete)
	copy(clone.Singles, a.Singles)
	return clone
}

// IsTing 是否听牌
func (a *Arrangement) IsTing() bool {
	// 牌型一: 2个不完整组合, 其余完整, 无单字
	if len(a.Incomplete) == 2 && len(a.Singles) == 0 {
		return true
	}
	// 牌型二: 1个单字, 其余完整, 无不完整组合
	if len(a.Incomplete) == 0 && len(a.Singles) == 1 {
		return true
	}
	return false
}

// TingType 听牌类型
type TingType int

const (
	TingTypeNone  TingType = iota
	TingTypeOne            // 2个不完整组合
	TingTypeTwo            // 1个单字
)

// GetTingType 获取听牌类型
func (a *Arrangement) GetTingType() TingType {
	if len(a.Incomplete) == 2 && len(a.Singles) == 0 {
		return TingTypeOne
	}
	if len(a.Incomplete) == 0 && len(a.Singles) == 1 {
		return TingTypeTwo
	}
	return TingTypeNone
}

// ArrangeTiles 理牌: 将一组牌组织为最优组合
// 考虑别字的多义性, 枚举所有别字分配方案后对每种方案做回溯搜索
func ArrangeTiles(tiles []*model.Tile, dangJing model.TileName) []*Arrangement {
	bieTiles := FilterBieTiles(tiles)
	assignments := EnumerateBieAssignments(bieTiles)

	bestHu := -1
	var bestArrangements []*Arrangement
	seen := make(map[string]bool) // 去重

	for _, assign := range assignments {
		counter := CountTilesWithBie(tiles, assign)
		bieMap := BuildBieMapping(assign)

		arrangements := backtrackArrange(counter, dangJing, bieMap, tiles)

		for _, arr := range arrangements {
			arr.TotalHu = calcArrangementHu(arr, dangJing)
			key := arrangementKey(arr)
			if seen[key] {
				continue
			}
			seen[key] = true

			if arr.TotalHu > bestHu {
				bestHu = arr.TotalHu
				bestArrangements = []*Arrangement{arr}
			} else if arr.TotalHu == bestHu {
				bestArrangements = append(bestArrangements, arr)
			}
		}
	}

	sort.Slice(bestArrangements, func(i, j int) bool {
		return bestArrangements[i].TotalHu > bestArrangements[j].TotalHu
	})

	return bestArrangements
}

// BestArrangement 找到胡数最大的排列方案
func BestArrangement(tiles []*model.Tile, dangJing model.TileName) *Arrangement {
	arrangements := ArrangeTiles(tiles, dangJing)
	if len(arrangements) == 0 {
		return NewArrangement()
	}
	return arrangements[0]
}

// backtrackArrange 回溯搜索所有合法组合
func backtrackArrange(counter map[model.TileName]int, dangJing model.TileName, bieMap map[int]model.TileName, originalTiles []*model.Tile) []*Arrangement {
	var results []*Arrangement
	current := NewArrangement()

	var search func(ctr map[model.TileName]int)
	search = func(ctr map[model.TileName]int) {
		if allUsed(ctr) {
			clone := current.Clone()
			// 收集剩余单字
			clone.Singles = collectSingles(ctr, bieMap, originalTiles)
			results = append(results, clone)
			return
		}

		// 尝试文字组合(完整)
		for _, word := range model.WordCombinations {
			if canFormWord(ctr, word) {
				useWord(ctr, word)
				comb := makeWordComb(word, model.CombComplete, bieMap, originalTiles)
				current.Complete = append(current.Complete, comb)
				search(ctr)
				current.Complete = current.Complete[:len(current.Complete)-1]
				unuseWord(ctr, word)
			}
		}

		// 尝试数字组合(完整)
		for _, nums := range model.NumericCombinations {
			if canFormNumeric(ctr, nums) {
				useNumeric(ctr, nums)
				comb := makeNumericComb(nums, model.CombComplete, bieMap, originalTiles)
				current.Complete = append(current.Complete, comb)
				search(ctr)
				current.Complete = current.Complete[:len(current.Complete)-1]
				unuseNumeric(ctr, nums)
			}
		}

		// 尝试单字组合(坎/统/五张统)
		for name, count := range ctr {
			if count >= 3 {
				useSingle(ctr, name, count)
				comb := makeSingleComb(name, count, model.CombComplete, bieMap, originalTiles)
				current.Complete = append(current.Complete, comb)
				search(ctr)
				current.Complete = current.Complete[:len(current.Complete)-1]
				unuseSingle(ctr, name, count)
			}
		}

		// 尝试不完整文字组合(2张)
		for _, word := range model.WordCombinations {
			for skip := 0; skip < 3; skip++ {
				if canFormWordIncomplete(ctr, word, skip) {
					useWordIncomplete(ctr, word, skip)
					comb := makeWordIncompleteComb(word, skip, bieMap, originalTiles)
					current.Incomplete = append(current.Incomplete, comb)
					search(ctr)
					current.Incomplete = current.Incomplete[:len(current.Incomplete)-1]
					unuseWordIncomplete(ctr, word, skip)
				}
			}
		}

		// 尝试不完整数字组合(2张)
		for _, nums := range model.NumericCombinations {
			for skip := 0; skip < 3; skip++ {
				if canFormNumericIncomplete(ctr, nums, skip) {
					useNumericIncomplete(ctr, nums, skip)
					comb := makeNumericIncompleteComb(nums, skip, bieMap, originalTiles)
					current.Incomplete = append(current.Incomplete, comb)
					search(ctr)
					current.Incomplete = current.Incomplete[:len(current.Incomplete)-1]
					unuseNumericIncomplete(ctr, nums, skip)
				}
			}
		}

		// 尝试对(2张相同)
		for name, count := range ctr {
			if count >= 2 {
				usePair(ctr, name)
				comb := makePairComb(name, bieMap, originalTiles)
				current.Incomplete = append(current.Incomplete, comb)
				search(ctr)
				current.Incomplete = current.Incomplete[:len(current.Incomplete)-1]
				unusePair(ctr, name)
			}
		}

		// 别字与经牌的不完整组合
		if ctr[model.TileBie] > 0 || hasJingInCounter(ctr) {
			searchBieIncomplete(ctr, bieMap, originalTiles, current, &results)
		}
	}

	search(counter)
	return results
}

// ===== 文字组合辅助函数 =====

func canFormWord(ctr map[model.TileName]int, word [3]model.TileName) bool {
	for _, name := range word {
		if ctr[name] <= 0 {
			return false
		}
	}
	return true
}

func useWord(ctr map[model.TileName]int, word [3]model.TileName) {
	for _, name := range word {
		ctr[name]--
		if ctr[name] <= 0 {
			delete(ctr, name)
		}
	}
}

func unuseWord(ctr map[model.TileName]int, word [3]model.TileName) {
	for _, name := range word {
		ctr[name]++
	}
}

func canFormWordIncomplete(ctr map[model.TileName]int, word [3]model.TileName, skip int) bool {
	for i, name := range word {
		if i == skip {
			continue
		}
		if ctr[name] <= 0 {
			return false
		}
	}
	return true
}

func useWordIncomplete(ctr map[model.TileName]int, word [3]model.TileName, skip int) {
	for i, name := range word {
		if i == skip {
			continue
		}
		ctr[name]--
		if ctr[name] <= 0 {
			delete(ctr, name)
		}
	}
}

func unuseWordIncomplete(ctr map[model.TileName]int, word [3]model.TileName, skip int) {
	for i, name := range word {
		if i == skip {
			continue
		}
		ctr[name]++
	}
}

// ===== 数字组合辅助函数 =====

func canFormNumeric(ctr map[model.TileName]int, nums [3]int) bool {
	for _, v := range nums {
		name := model.NumericNameFromValue(v)
		if ctr[name] <= 0 {
			return false
		}
	}
	return true
}

func useNumeric(ctr map[model.TileName]int, nums [3]int) {
	for _, v := range nums {
		name := model.NumericNameFromValue(v)
		ctr[name]--
		if ctr[name] <= 0 {
			delete(ctr, name)
		}
	}
}

func unuseNumeric(ctr map[model.TileName]int, nums [3]int) {
	for _, v := range nums {
		name := model.NumericNameFromValue(v)
		ctr[name]++
	}
}

func canFormNumericIncomplete(ctr map[model.TileName]int, nums [3]int, skip int) bool {
	for i, v := range nums {
		if i == skip {
			continue
		}
		name := model.NumericNameFromValue(v)
		if ctr[name] <= 0 {
			return false
		}
	}
	return true
}

func useNumericIncomplete(ctr map[model.TileName]int, nums [3]int, skip int) {
	for i, v := range nums {
		if i == skip {
			continue
		}
		name := model.NumericNameFromValue(v)
		ctr[name]--
		if ctr[name] <= 0 {
			delete(ctr, name)
		}
	}
}

func unuseNumericIncomplete(ctr map[model.TileName]int, nums [3]int, skip int) {
	for i, v := range nums {
		if i == skip {
			continue
		}
		name := model.NumericNameFromValue(v)
		ctr[name]++
	}
}

// ===== 单字组合辅助函数 =====

func useSingle(ctr map[model.TileName]int, name model.TileName, count int) {
	delete(ctr, name)
}

func unuseSingle(ctr map[model.TileName]int, name model.TileName, count int) {
	ctr[name] = count
}

func usePair(ctr map[model.TileName]int, name model.TileName) {
	ctr[name] -= 2
	if ctr[name] <= 0 {
		delete(ctr, name)
	}
}

func unusePair(ctr map[model.TileName]int, name model.TileName) {
	ctr[name] += 2
}

// ===== 别字不完整组合搜索 =====

func searchBieIncomplete(ctr map[model.TileName]int, bieMap map[int]model.TileName, originalTiles []*model.Tile, current *Arrangement, results *[]*Arrangement) {
	// 别字与一张经牌/数字牌的不完整组合
	bieCount := ctr[model.TileBie]
	if bieCount <= 0 {
		return
	}

	// "别" 和任何1张"乙二三四五六七八九十土化千"均可组成不完整组合
	// 此处简化: 别字参与的不完整组合由不完整数字/文字组合已经覆盖
	// 额外处理: 别字与1张非经牌组成的不完整组合(如"五别"视作不完整坎)
	for name := range ctr {
		if name == model.TileBie {
			continue
		}
		if model.IsJingName(name) {
			// 别字+1张经牌 = 不完整坎(别视作该经牌)
			ctr[model.TileBie]--
			if ctr[model.TileBie] <= 0 {
				delete(ctr, model.TileBie)
			}
			ctr[name]--
			if ctr[name] <= 0 {
				delete(ctr, name)
			}
			comb := model.NewCombination(model.CombSingle, model.CombIncomplete, nil)
			current.Incomplete = append(current.Incomplete, comb)
			searchSimple(ctr, current, results)
			current.Incomplete = current.Incomplete[:len(current.Incomplete)-1]
			ctr[name]++
			ctr[model.TileBie]++
		}
	}
}

func searchSimple(ctr map[model.TileName]int, current *Arrangement, results *[]*Arrangement) {
	if allUsed(ctr) {
		clone := current.Clone()
		*results = append(*results, clone)
	}
}

// ===== 组合构建辅助函数 =====

func makeWordComb(word [3]model.TileName, completeness model.CombCompleteness, bieMap map[int]model.TileName, originalTiles []*model.Tile) *model.Combination {
	comb := model.NewCombination(model.CombWord, completeness, nil)
	for _, name := range word {
		tile := findTileByName(originalTiles, name, comb)
		if tile != nil {
			comb.Tiles = append(comb.Tiles, tile)
		}
	}
	return comb
}

func makeNumericComb(nums [3]int, completeness model.CombCompleteness, bieMap map[int]model.TileName, originalTiles []*model.Tile) *model.Combination {
	comb := model.NewCombination(model.CombNumeric, completeness, nil)
	for _, v := range nums {
		name := model.NumericNameFromValue(v)
		tile := findTileByName(originalTiles, name, comb)
		if tile != nil {
			comb.Tiles = append(comb.Tiles, tile)
		}
	}
	return comb
}

func makeSingleComb(name model.TileName, count int, completeness model.CombCompleteness, bieMap map[int]model.TileName, originalTiles []*model.Tile) *model.Combination {
	comb := model.NewCombination(model.CombSingle, completeness, nil)
	for i := 0; i < count; i++ {
		tile := findTileByName(originalTiles, name, comb)
		if tile != nil {
			comb.Tiles = append(comb.Tiles, tile)
		}
	}
	comb.BieMapping = bieMap
	return comb
}

func makeWordIncompleteComb(word [3]model.TileName, skip int, bieMap map[int]model.TileName, originalTiles []*model.Tile) *model.Combination {
	comb := model.NewCombination(model.CombWord, model.CombIncomplete, nil)
	for i, name := range word {
		if i == skip {
			continue
		}
		tile := findTileByName(originalTiles, name, comb)
		if tile != nil {
			comb.Tiles = append(comb.Tiles, tile)
		}
	}
	return comb
}

func makeNumericIncompleteComb(nums [3]int, skip int, bieMap map[int]model.TileName, originalTiles []*model.Tile) *model.Combination {
	comb := model.NewCombination(model.CombNumeric, model.CombIncomplete, nil)
	for i, v := range nums {
		if i == skip {
			continue
		}
		name := model.NumericNameFromValue(v)
		tile := findTileByName(originalTiles, name, comb)
		if tile != nil {
			comb.Tiles = append(comb.Tiles, tile)
		}
	}
	return comb
}

func makePairComb(name model.TileName, bieMap map[int]model.TileName, originalTiles []*model.Tile) *model.Combination {
	comb := model.NewCombination(model.CombSingle, model.CombIncomplete, nil)
	for i := 0; i < 2; i++ {
		tile := findTileByName(originalTiles, name, comb)
		if tile != nil {
			comb.Tiles = append(comb.Tiles, tile)
		}
	}
	return comb
}

// findTileByName 从原始牌中查找指定名称的牌(排除已在组合中的)
func findTileByName(tiles []*model.Tile, name model.TileName, currentComb *model.Combination) *model.Tile {
	usedIDs := make(map[int]bool)
	if currentComb != nil {
		for _, t := range currentComb.Tiles {
			usedIDs[t.ID] = true
		}
	}
	for _, t := range tiles {
		if !usedIDs[t.ID] {
			resolved := t.Name
			if t.Name == model.TileBie {
				// 别字: 检查是否映射为该名称
				if mapped, ok := currentComb.BieMapping[t.ID]; ok {
					resolved = mapped
				}
			}
			if resolved == name {
				return t
			}
		}
	}
	return nil
}

// ===== 通用辅助 =====

func allUsed(ctr map[model.TileName]int) bool {
	for _, v := range ctr {
		if v > 0 {
			return false
		}
	}
	return true
}

func collectSingles(ctr map[model.TileName]int, bieMap map[int]model.TileName, originalTiles []*model.Tile) []*model.Tile {
	var singles []*model.Tile
	for name, count := range ctr {
		for i := 0; i < count; i++ {
			tile := findTileByName(originalTiles, name, nil)
			if tile != nil {
				singles = append(singles, tile)
			}
		}
	}
	return singles
}

func hasJingInCounter(ctr map[model.TileName]int) bool {
	return ctr[model.TileSan] > 0 || ctr[model.TileWu] > 0 || ctr[model.TileQi] > 0
}

func calcArrangementHu(arr *Arrangement, dangJing model.TileName) int {
	total := 0
	for _, comb := range arr.Complete {
		total += CalcCombHu(comb, dangJing)
	}
	for _, comb := range arr.Incomplete {
		total += calcIncompleteHu(comb, dangJing)
	}
	for _, tile := range arr.Singles {
		resolved := tile.Name
		total += CalcSingleTileHu(tile, resolved, dangJing)
	}
	return total
}

// calcIncompleteHu 不完整组合的胡数
// 不完整组合本身0胡, 需从单字本身算胡
func calcIncompleteHu(comb *model.Combination, dangJing model.TileName) int {
	total := 0
	for _, tile := range comb.Tiles {
		resolved := comb.ResolveName(tile)
		total += CalcSingleTileHu(tile, resolved, dangJing)
	}
	return total
}

func arrangementKey(arr *Arrangement) string {
	// 简单去重: 用完整组合数量+不完整组合数量+单字数量作为key
	// 实际需要更精确的去重, 但用于初步实现
	return fmt.Sprintf("%d-%d-%d", len(arr.Complete), len(arr.Incomplete), len(arr.Singles))
}
