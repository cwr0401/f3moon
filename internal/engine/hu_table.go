package engine

import "github.com/cwr0401/f3moon/internal/model"

// HuPair 当经/不当经胡数对
type HuPair struct {
	Dang    int // 当经胡数
	BuDang  int // 不当经胡数
}

// JingKey 经牌组合查询键
type JingKey struct {
	Size   int // 3=坎, 4=统, 5=五张统
	Flower int // 花经数量
	Bie    int // 别字数量
}

// JingHuTable 经牌组合胡数表
// key: {总张数, 花经数, 别字数}, value: {当经胡数, 不当经胡数}
// 白经数 = Size - Flower - Bie
var JingHuTable = map[JingKey]HuPair{
	// ===== 坎(3张) =====
	{Size: 3, Flower: 0, Bie: 0}: {Dang: 10, BuDang: 5},   // 3白经
	{Size: 3, Flower: 1, Bie: 0}: {Dang: 12, BuDang: 6},   // 2白+1花
	{Size: 3, Flower: 0, Bie: 1}: {Dang: 12, BuDang: 6},   // 2白+1别
	{Size: 3, Flower: 2, Bie: 0}: {Dang: 14, BuDang: 7},   // 1白+2花
	{Size: 3, Flower: 1, Bie: 1}: {Dang: 14, BuDang: 7},   // 1白+1花+1别
	{Size: 3, Flower: 0, Bie: 2}: {Dang: 14, BuDang: 7},   // 1白+2别
	{Size: 3, Flower: 2, Bie: 1}: {Dang: 18, BuDang: 9},   // 2花+1别
	{Size: 3, Flower: 1, Bie: 2}: {Dang: 24, BuDang: 12},  // 1花+2别

	// ===== 统(4张) =====
	{Size: 4, Flower: 0, Bie: 0}: {Dang: 20, BuDang: 10},  // 4白经
	{Size: 4, Flower: 1, Bie: 0}: {Dang: 24, BuDang: 12},  // 3白+1花
	{Size: 4, Flower: 0, Bie: 1}: {Dang: 24, BuDang: 12},  // 3白+1别
	{Size: 4, Flower: 2, Bie: 0}: {Dang: 28, BuDang: 14},  // 2白+2花
	{Size: 4, Flower: 1, Bie: 1}: {Dang: 28, BuDang: 14},  // 2白+1花+1别
	{Size: 4, Flower: 0, Bie: 2}: {Dang: 28, BuDang: 14},  // 2白+2别
	{Size: 4, Flower: 2, Bie: 1}: {Dang: 36, BuDang: 18},  // 1白+2花+1别
	{Size: 4, Flower: 1, Bie: 2}: {Dang: 36, BuDang: 18},  // 1白+1花+2别
	{Size: 4, Flower: 2, Bie: 2}: {Dang: 48, BuDang: 24},  // 2花+2别

	// ===== 五张统(5张) =====
	{Size: 5, Flower: 0, Bie: 0}: {Dang: 40, BuDang: 20},  // 5白经(理论值)
	{Size: 5, Flower: 2, Bie: 0}: {Dang: 56, BuDang: 28},  // 3白+2花
	{Size: 5, Flower: 0, Bie: 2}: {Dang: 56, BuDang: 28},  // 3白+2别
	{Size: 5, Flower: 1, Bie: 1}: {Dang: 56, BuDang: 28},  // 3白+1花+1别
	{Size: 5, Flower: 2, Bie: 1}: {Dang: 72, BuDang: 36},  // 2白+2花+1别
	{Size: 5, Flower: 1, Bie: 2}: {Dang: 72, BuDang: 36},  // 2白+1花+2别
	{Size: 5, Flower: 2, Bie: 2}: {Dang: 96, BuDang: 48},  // 1白+2花+2别
}

// LookupJingHu 查询经牌组合胡数
func LookupJingHu(size, flower, bie int) HuPair {
	key := JingKey{Size: size, Flower: flower, Bie: bie}
	if pair, ok := JingHuTable[key]; ok {
		return pair
	}
	// 回退: 用通用公式计算
	return calcJingHuFallback(size, flower, bie)
}

// calcJingHuFallback 经牌组合胡数的通用推导公式
// 规则: 坎3白经=10/5, 每增加1花经+2/+1, 每增加1别字+2/+1(在花经基础上额外叠加)
// 统 = 坎 × 2, 五张统 = 坎 × 4
func calcJingHuFallback(size, flower, bie int) HuPair {
	// 基础: 坎3白经
	baseDang := 10
	baseBuDang := 5

	// 每增加1张(>3张), 相当于在坎的基础上翻倍扩展
	// 简化: 用已有表格值做线性外推
	extraFlower := flower
	extraBie := bie

	// 花经加成: 每花+2(当经)/+1(不当经)
	// 别字加成: 每别+2(当经)/+1(不当经) — 别字等同于花经
	dang := baseDang + extraFlower*2 + extraBie*2
	buDang := baseBuDang + extraFlower + extraBie

	// 张数倍率: 坎×1, 统×2, 五张统×4
	multiplier := 1
	switch size {
	case 4:
		multiplier = 2
	case 5:
		multiplier = 4
	}

	return HuPair{
		Dang:   dang * multiplier,
		BuDang: buDang * multiplier,
	}
}

// CalcOpenCombHu 计算明牌区碰牌组合(对/招/泛)的胡数
func CalcOpenCombHu(comb *model.Combination, dangJing model.TileName) int {
	size := len(comb.Tiles)
	resolved := comb.ResolveName(comb.Tiles[0])

	if model.IsJingName(resolved) {
		// 经牌碰牌组合: 逐张计算
		totalHu := 0
		for _, t := range comb.Tiles {
			r := comb.ResolveName(t)
			isDang := (r == dangJing)
			if t.Name == model.TileBie {
				// 别字视作花经
				if isDang {
					totalHu += 4
				} else {
					totalHu += 2
				}
			} else if t.IsFlower {
				if isDang {
					totalHu += 4
				} else {
					totalHu += 2
				}
			} else {
				// 白经
				if isDang {
					totalHu += 2
				} else {
					totalHu += 1
				}
			}
		}
		return totalHu
	}

	// 非经牌
	switch size {
	case 3: // 对牌形成的坎
		if comb.HasRedTile() {
			return 1 // 红色一坎1胡
		}
		// 黑色坎: 0胡 + 花加成
		return comb.FlowerCount()
	case 4: // 招牌形成的统
		return CalcCombHu(comb, dangJing)
	case 5: // 泛牌形成的五张统
		return CalcCombHu(comb, dangJing)
	}
	return 0
}
