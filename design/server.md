# 荆楚花牌 - 服务端设计

## 1. 项目概述

基于荆楚花牌传统规则，设计一个支持多人在线对战的纸牌游戏服务端。采用 Go 语言实现，使用 Gin 框架提供 HTTP/WebSocket 服务。

### 核心特性

- 4人一桌（3人打牌 + 1人歇家），支持3人定庄模式
- 完整的牌局流程：洗牌 → 切牌 → 起牌 → 请统 → 打牌 → 和牌/黄庄
- 精确的胡数计算与听牌检测
- "别"字牌的多义性处理
- AI 玩家支持（基于博弈策略）
- WebSocket 实时通信

---

## 2. 架构总览

```
┌─────────────────────────────────────────────────┐
│                   Client                        │
│         (Web/H5/小程序/桌面端)                    │
└────────────────────┬────────────────────────────┘
                     │ WebSocket / HTTP
┌────────────────────▼────────────────────────────┐
│              API Gateway (Gin)                   │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ Room API │  │ Game API │  │ WebSocket Hub │  │
│  └──────────┘  └──────────┘  └───────────────┘  │
├─────────────────────────────────────────────────┤
│              Service Layer                       │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌────────────┐  │
│  │Room  │  │Game  │  │Hu    │  │AI          │  │
│  │Svc   │  │Svc   │  │Calc  │  │Player      │  │
│  └──────┘  └──────┘  └──────┘  └────────────┘  │
├─────────────────────────────────────────────────┤
│              Core Engine                         │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌────────────┐  │
│  │Card  │  │Comb  │  │State │  │Shuffle     │  │
│  │Model │  │Detect│  │Machine│ │& Deal      │  │
│  └──────┘  └──────┘  └──────┘  └────────────┘  │
├─────────────────────────────────────────────────┤
│              Infra                               │
│  ┌──────┐  ┌──────┐  ┌──────────────────────┐  │
│  │Redis │  │Log   │  │Config                │  │
│  │(可选) │  │      │  │                      │  │
│  └──────┘  └──────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────┘
```

---

## 3. 数据模型

### 3.1 牌（Tile）

```go
// TileName 牌面名称
type TileName string

const (
    // 数字牌 1-10
    TileYi   TileName = "乙"  // 1
    TileEr   TileName = "二"  // 2
    TileSan  TileName = "三"  // 3
    TileSi   TileName = "四"  // 4
    TileWu   TileName = "五"  // 5
    TileLiu  TileName = "六"  // 6
    TileQi   TileName = "七"  // 7
    TileBa   TileName = "八"  // 8
    TileJiu  TileName = "九"  // 9
    TileShi  TileName = "十"  // 10
    // 文字牌
    TileKong TileName = "孔"
    TileJi   TileName = "己"
    TileHua  TileName = "化"
    TileQian TileName = "千"
    TileTu   TileName = "土"
    TileZi   TileName = "子"
    TileShang TileName = "上"
    TileDa   TileName = "大"
    TileRen  TileName = "人"
    TileKe   TileName = "可"
    TileZhi  TileName = "知"
    TileLi   TileName = "礼"
    // 特殊牌
    TileBie  TileName = "别"
)

// TileColor 牌色
type TileColor int

const (
    ColorBlack TileColor = iota
    ColorRed
)

// Tile 一张牌的完整属性
type Tile struct {
    ID    int       `json:"id"`     // 唯一标识 0-111
    Name  TileName  `json:"name"`  // 牌面字
    Color TileColor `json:"color"` // 颜色
    IsFlower bool   `json:"is_flower"` // 是否带花
    Numeric  int    `json:"numeric"`   // 数字值(仅数字牌有效, 1-10; 文字牌/别=0)
}

// IsJing 是否为经牌(三、五、七)
func (t *Tile) IsJing() bool {
    return t.Name == TileSan || t.Name == TileWu || t.Name == TileQi
}
```

### 3.2 牌组初始化

112 张牌的生成规则：

| 牌名 | 颜色 | 张数 | 其中带花 | 其中不带花 |
|------|------|------|----------|------------|
| 乙 | 黑 | 5 | 2 | 3 |
| 二 | 黑 | 5 | 0 | 5 |
| 三 | 红 | 5 | 2 | 3 |
| 四 | 黑 | 5 | 0 | 5 |
| 五 | 红 | 5 | 2 | 3 |
| 六 | 黑 | 5 | 0 | 5 |
| 七 | 红 | 5 | 2 | 3 |
| 八 | 黑 | 5 | 0 | 5 |
| 九 | 黑 | 5 | 2 | 3 |
| 十 | 黑 | 5 | 0 | 5 |
| 孔 | 黑 | 5 | 0 | 5 |
| 己 | 黑 | 5 | 0 | 5 |
| 化 | 黑 | 5 | 0 | 5 |
| 千 | 黑 | 5 | 0 | 5 |
| 土 | 黑 | 5 | 0 | 5 |
| 子 | 黑 | 5 | 0 | 5 |
| 上 | 红 | 5 | 0 | 5 |
| 大 | 红 | 5 | 0 | 5 |
| 人 | 红 | 5 | 0 | 5 |
| 可 | 红 | 5 | 0 | 5 |
| 知 | 红 | 5 | 0 | 5 |
| 礼 | 红 | 5 | 0 | 5 |
| 别 | 红 | 2 | 2(始终带花) | 0 |

### 3.3 组合（Combination）

```go
// CombType 组合类型
type CombType int

const (
    CombWord      CombType = iota // 文字组合(3张一句)
    CombNumeric                   // 数字组合(3张连续一句)
    CombSingle                    // 单字组合(坎/统/五张统)
)

// CombCompleteness 组合完整度
type CombCompleteness int

const (
    CombComplete   CombCompleteness = iota // 完整组合
    CombIncomplete                         // 不完整组合(2张)
)

// Combination 一个牌组合
type Combination struct {
    Type         CombType        `json:"type"`
    Completeness CombCompleteness `json:"completeness"`
    Tiles        []*Tile         `json:"tiles"`        // 实际使用的牌
    BieMapping   map[int]TileName `json:"bie_mapping"` // 别字映射: tileID -> 别字视作的牌名
}
```

### 3.4 玩家（Player）

```go
type PlayerRole int

const (
    RoleDealer  PlayerRole = iota // 庄家
    RoleIdle1                     // 闲一(庄家右手边)
    RoleIdle2                     // 闲二(庄家左手边)
    RoleRest                      // 歇家
)

type Player struct {
    ID          string        `json:"id"`
    Role        PlayerRole    `json:"role"`
    IsAI        bool          `json:"is_ai"`

    Hand        []*Tile       `json:"hand"`          // 手牌(仅自己可见)
    OpenTiles   []*Tile       `json:"open_tiles"`    // 明牌区(公开)
    OpenCombs   []*Combination `json:"open_combs"`   // 明牌区组合(公开)

    DangJing    TileName      `json:"dang_jing"`     // 当经(三/五/七)
    PairCount   int           `json:"pair_count"`    // 已对牌次数(上限2)
}
```

### 3.5 牌局状态（GameState）

```go
type GamePhase int

const (
    PhaseWaiting    GamePhase = iota // 等待玩家就位
    PhaseShuffle                     // 洗牌
    PhaseCut                         // 切牌(腰牌)
    PhaseDeal                        // 起牌
    PhaseTongAsk                     // 请统
    PhasePlay                        // 打牌
    PhaseCheck                       // 和牌查验
    PhaseFinished                    // 牌局结束
)

type TurnAction int

const (
    ActionDraw    TurnAction = iota // 起牌
    ActionDiscard                    // 出牌
    ActionPair                       // 对
    ActionZhao                       // 招
    ActionFan                        // 泛
    ActionGanta                      // 赶塔
    ActionTong                       // 统牌
    ActionWin                        // 和牌
    ActionPass                       // 过
)

type GameState struct {
    ID            string       `json:"id"`
    Phase         GamePhase    `json:"phase"`
    Mode          GameMode     `json:"mode"`           // 4人/3人

    Players       [4]*Player   `json:"players"`        // 0=庄, 1=闲一, 2=闲二, 3=歇家(3人模式为nil)
    DrawPile      []*Tile      `json:"draw_pile"`      // 公牌
    DiscardPile   []*Tile      `json:"discard_pile"`   // 出牌区

    CurrentTurn   int          `json:"current_turn"`   // 当前出牌玩家索引
    DealerIndex   int          `json:"dealer_index"`   // 庄家索引

    TongRequested bool         `json:"tong_requested"` // 是否已请统
    TongOrder     []int        `json:"tong_order"`     // 统牌抉择顺序
    TongCurrent   int          `json:"tong_current"`   // 当前统牌抉择人

    LastDiscard   *Tile        `json:"last_discard"`   // 最后打出的牌
    LastDiscarder int          `json:"last_discarder"` // 最后出牌者索引

    Winner        int          `json:"winner"`         // 赢家索引(-1表示未定)
    WinType       WinType      `json:"win_type"`       // 自摸/点炮/海底捞月
}
```

---

## 4. 核心算法设计

### 4.1 胡数计算

胡数计算是系统最核心也最复杂的算法。采用**分层累加**策略，将规则系统化为公式。

#### 4.1.1 基础胡数表

```go
// HuBase 基础胡数(不含花、经加成)
var HuBase = map[CombKey]int{
    // 一句(3张)
    {Type: CombWord, Color: ColorBlack, Completeness: CombComplete}:     0,
    {Type: CombWord, Color: ColorRed,   Completeness: CombComplete}:     1,
    {Type: CombNumeric, Completeness: CombComplete}:                     0,
    // 坎(3张相同)
    {Type: CombSingle, Color: ColorBlack, SingleSize: 3}:               1,
    {Type: CombSingle, Color: ColorRed,   SingleSize: 3}:               2,
    // 统(4张相同)
    {Type: CombSingle, Color: ColorBlack, SingleSize: 4}:               2,
    {Type: CombSingle, Color: ColorRed,   SingleSize: 4}:               4,
    // 五张统(5张相同)
    {Type: CombSingle, Color: ColorBlack, SingleSize: 5}:               4,
    {Type: CombSingle, Color: ColorRed,   SingleSize: 5}:               8,
}
```

#### 4.1.2 加成规则

胡数 = 基础胡 + 花加成 + 经加成

```
花加成:
  非经牌: 每花 +1
  经牌(不当经): 每花 +1 (在单字/坎/统基础上)
  经牌(当经): 每花 +2 (在单字/坎/统基础上)

经加成(仅经牌 三/五/七):
  当经倍率: ×2
  不当经倍率: ×1
  即: 经牌部分的胡数 × 当经倍率

特别地，经牌的坎/统/五张统有独立的胡数表，不按简单倍率计算
```

#### 4.1.3 经牌组合胡数表

经牌组合（坎/统/五张统）胡数不遵循简单倍率，需查表：

```go
// JingHuTable 经牌组合胡数表
// key: {单字数, 花经数, 别字数}, value: {当经胡数, 不当经胡数}
// 白经=单字数-花经数-别字数
var JingHuTable = map[JingKey]HuPair{
    // 坎(3张)
    {Size: 3, Flower: 0, Bie: 0}: {Dang: 10, BuDang: 5},    // 3白经
    {Size: 3, Flower: 1, Bie: 0}: {Dang: 12, BuDang: 6},    // 2白+1花
    {Size: 3, Flower: 0, Bie: 1}: {Dang: 12, BuDang: 6},    // 2白+1别
    {Size: 3, Flower: 2, Bie: 0}: {Dang: 14, BuDang: 7},    // 1白+2花
    {Size: 3, Flower: 1, Bie: 1}: {Dang: 14, BuDang: 7},    // 1白+1花+1别
    {Size: 3, Flower: 0, Bie: 2}: {Dang: 14, BuDang: 7},    // 1白+2别
    {Size: 3, Flower: 2, Bie: 1}: {Dang: 18, BuDang: 9},    // 2花+1别
    {Size: 3, Flower: 1, Bie: 2}: {Dang: 24, BuDang: 12},   // 1花+2别
    // 统(4张)
    {Size: 4, Flower: 0, Bie: 0}: {Dang: 20, BuDang: 10},   // 4白经
    {Size: 4, Flower: 1, Bie: 0}: {Dang: 24, BuDang: 12},   // 3白+1花
    {Size: 4, Flower: 0, Bie: 1}: {Dang: 24, BuDang: 12},   // 3白+1别
    {Size: 4, Flower: 2, Bie: 0}: {Dang: 28, BuDang: 14},   // 2白+2花
    {Size: 4, Flower: 1, Bie: 1}: {Dang: 28, BuDang: 14},   // 2白+1花+1别
    {Size: 4, Flower: 0, Bie: 2}: {Dang: 28, BuDang: 14},   // 2白+2别
    {Size: 4, Flower: 2, Bie: 1}: {Dang: 36, BuDang: 18},   // 1白+2花+1别
    {Size: 4, Flower: 1, Bie: 2}: {Dang: 36, BuDang: 18},   // 1白+1花+2别
    {Size: 4, Flower: 2, Bie: 2}: {Dang: 48, BuDang: 24},   // 2花+2别
    // 五张统(5张)
    {Size: 5, Flower: 0, Bie: 0}: {Dang: 40, BuDang: 20},   // 5白经
    {Size: 5, Flower: 2, Bie: 0}: {Dang: 56, BuDang: 28},   // 3白+2花
    {Size: 5, Flower: 0, Bie: 2}: {Dang: 56, BuDang: 28},   // 3白+2别
    {Size: 5, Flower: 1, Bie: 1}: {Dang: 56, BuDang: 28},   // 3白+1花+1别
    {Size: 5, Flower: 2, Bie: 1}: {Dang: 72, BuDang: 36},   // 2白+2花+1别
    {Size: 5, Flower: 1, Bie: 2}: {Dang: 72, BuDang: 36},   // 2白+1花+2别
    {Size: 5, Flower: 2, Bie: 2}: {Dang: 96, BuDang: 48},   // 1白+2花+2别
}
```

> **注意**: 规则文档中未列举所有经牌组合的胡数（如5白经的坎、4白经的统），上表中缺失的条目需要根据规则推导补充。推导方法：经牌的5张牌中必有2张带花，因此"5白经"等不含花的5张统在实际中不可能出现，但算法仍需覆盖边界情况。建议采用**通用公式**计算。

#### 4.1.4 通用胡数计算公式

```go
// CalcHu 计算一个组合的胡数
// 原则: 基础胡 + 花加成, 经牌部分单独计算后合并
func CalcHu(comb *Combination, dangJing TileName) int {
    switch comb.Type {
    case CombWord, CombNumeric:
        return calcSentenceHu(comb, dangJing)
    case CombSingle:
        return calcSingleHu(comb, dangJing)
    }
    return 0
}

// 一句(文字/数字组合)的胡数计算
// = 基础胡(黑0/红1) + 花加成(非经每花+1) + 经牌加成
func calcSentenceHu(comb *Combination, dangJing TileName) int {
    baseHu := 0
    if comb.hasRedTile() {
        baseHu = 1 // 红色一句
    }

    flowerBonus := 0
    jingHu := 0
    for _, tile := range comb.Tiles {
        actualName := resolveBie(tile, comb.BieMapping)
        if isJing(actualName) {
            jingHu += calcJingTileHu(tile, actualName, dangJing)
        } else if tile.IsFlower {
            flowerBonus++
        }
    }

    return baseHu + flowerBonus + jingHu
}

// 经牌单字胡数
// 白经: 当经2胡, 不当经1胡
// 花经: 当经4胡, 不当经2胡
func calcJingTileHu(tile *Tile, actualName TileName, dangJing TileName) int {
    isDang := (actualName == dangJing)
    if tile.IsFlower || tile.Name == TileBie {
        if isDang { return 4 } else { return 2 }
    }
    if isDang { return 2 } else { return 1 }
}
```

### 4.2 组合检测算法

理牌的核心问题：给定一组牌，找到所有合法的组合方式。

#### 4.2.1 别字处理策略

"别"字是最复杂的部分，它可以在不同组合中视作不同的字。采用**回溯搜索 + 剪枝**策略：

```go
// BieAssignment 别字分配方案
type BieAssignment struct {
    TileID int
    AsName TileName // 别视作的牌名(三/五/七)
}

// ArrangeTiles 理牌: 将一组牌组织为最优组合
// 返回所有可能的排列方案, 按胡数降序排列
func ArrangeTiles(tiles []*Tile, dangJing TileName) []*Arrangement {
    // 1. 识别所有别字
    bieTiles := filterBie(tiles)

    // 2. 枚举别字的所有可能分配(2^3种: 每张别可作花三/花五/花七)
    bieAssignments := enumerateBieAssignments(bieTiles)

    // 3. 对每种分配方案, 进行回溯搜索找最优组合
    var bestArrangements []*Arrangement
    for _, assignment := range bieAssignments {
        resolved := resolveBieAssignments(tiles, assignment)
        arrangements := backtrackArrange(resolved, dangJing)
        bestArrangements = append(bestArrangements, arrangements...)
    }

    // 4. 按胡数降序排列
    sort.Slice(bestArrangements, func(i, j int) bool {
        return bestArrangements[i].TotalHu > bestArrangements[j].TotalHu
    })

    return bestArrangements
}
```

#### 4.2.2 回溯理牌算法

```go
// backtrackArrange 回溯搜索所有合法组合
// 优先级: 完整组合 > 不完整组合 > 单字
func backtrackArrange(tiles []*Tile, dangJing TileName) []*Arrangement {
    counter := buildCounter(tiles) // 统计每种牌的数量

    var results []*Arrangement
    var current Arrangement

    var search func(counter map[TileName]int)
    search = func(counter map[TileName]int) {
        if allUsed(counter) {
            results = append(results, current.Clone())
            return
        }

        // 尝试文字组合
        for _, word := range WordCombinations {
            if canFormWord(counter, word) {
                useWord(counter, word)
                current.AddCompleteComb(word, CombWord)
                search(counter)
                current.RemoveLast()
                unuseWord(counter, word)
            }
        }

        // 尝试数字组合
        for start := 1; start <= 8; start++ {
            if canFormNumeric(counter, start) {
                useNumeric(counter, start)
                current.AddCompleteComb(numericComb(start), CombNumeric)
                search(counter)
                current.RemoveLast()
                unuseNumeric(counter, start)
            }
        }

        // 尝试单字组合(坎/统/五张统)
        for name, count := range counter {
            if count >= 3 {
                useSingle(counter, name, count)
                current.AddCompleteComb(singleComb(name, count), CombSingle)
                search(counter)
                current.RemoveLast()
                unuseSingle(counter, name, count)
            }
        }

        // 尝试不完整组合(2张)
        // ... 文字不完整、数字不完整、对
    }

    search(counter)
    return results
}
```

#### 4.2.3 剪枝优化

回溯搜索的搜索空间可能很大，需采用以下剪枝策略：

1. **剩余牌数检查**：剩余牌无法形成任何合法组合时提前剪枝
2. **胡数上界剪枝**：当前胡数 + 剩余牌理论最大胡数 < 17 时剪枝
3. **贪心初始解**：先用贪心策略找到一个可行解，后续搜索中若当前胡数已低于最优解则剪枝
4. **等价方案去重**：同一牌型存在多种等价排列时只保留一种

### 4.3 听牌检测

```go
// TingResult 听牌检测结果
type TingResult struct {
    Arrangement *Arrangement  // 听牌时的牌型排列
    TingTiles   []TileName    // 听字列表
    TingType    TingType      // 听牌牌型(一/二)
    TotalHu     int           // 当前总胡数
}

// DetectTing 检测手牌是否听牌
// 返回所有听牌方案
func DetectTing(hand []*Tile, openCombs []*Combination, dangJing TileName) []*TingResult {
    arrangements := ArrangeTiles(hand, dangJing)
    var results []*TingResult

    for _, arr := range arrangements {
        // 牌型一: 2个不完整组合, 其余完整
        if len(arr.Incomplete) == 2 && len(arr.Singles) == 0 {
            tingTiles := calcTingTiles(arr, dangJing)
            if len(tingTiles) >= 2 {
                results = append(results, &TingResult{
                    Arrangement: arr,
                    TingTiles:   tingTiles,
                    TingType:    TingTypeOne,
                })
            }
        }
        // 牌型二: 1个单字, 其余完整
        if len(arr.Incomplete) == 0 && len(arr.Singles) == 1 {
            tingTiles := calcTingTilesFromSingle(arr, dangJing)
            if len(tingTiles) >= 2 {
                results = append(results, &TingResult{
                    Arrangement: arr,
                    TingTiles:   tingTiles,
                    TingType:    TingTypeTwo,
                })
            }
        }
    }

    return results
}

// calcTingTiles 计算听字
// 对于每个不完整组合, 找出能补全的牌
// 对于单字, 找出能与单字组成不完整组合的牌
func calcTingTiles(arr *Arrangement, dangJing TileName) []TileName {
    var tingTiles []TileName
    tingSet := make(map[TileName]bool)

    for _, inc := range arr.Incomplete {
        completions := findCompletions(inc, dangJing)
        for _, t := range completions {
            tingSet[t] = true
        }
    }

    for name := range tingSet {
        tingTiles = append(tingTiles, name)
    }

    // 别字作为听字: 如果听字含三/五/七, 别也是听字
    for _, t := range tingTiles {
        if t == TileSan || t == TileWu || t == TileQi {
            tingSet[TileBie] = true
        }
    }

    return tingTiles
}
```

### 4.4 当经选择算法

```go
// ChooseDangJing 选择最优当经
// 枚举三/五/七, 计算每种选择下的总胡数, 取最大
func ChooseDangJing(hand []*Tile, openCombs []*Combination) TileName {
    bestJing := TileSan
    bestHu := -1

    for _, jing := range []TileName{TileSan, TileWu, TileQi} {
        arrangement := BestArrangement(hand, jing)
        totalHu := arrangement.TotalHu + calcOpenHu(openCombs, jing)
        if totalHu > bestHu {
            bestHu = totalHu
            bestJing = jing
        }
    }

    return bestJing
}
```

---

## 5. 游戏状态机

### 5.1 状态流转

```
┌─────────┐  4人就位   ┌─────────┐  庄家洗牌  ┌─────────┐
│ Waiting │──────────►│ Shuffle │──────────►│   Cut   │
└─────────┘          └─────────┘          └────┬────┘
                                                │ 歇家/闲二切牌
                                           ┌────▼────┐
                                           │  Deal   │
                                           │起牌25张 │
                                           └────┬────┘
                                                │ 庄家起第26张
                                           ┌────▼────┐
                                           │TongAsk  │
                                           │请统阶段 │
                                           └────┬────┘
                                                │ 三人抉择完毕
                                           ┌────▼────┐
                                           │  Play   │◄──────────┐
                                           │  打牌   │            │
                                           └────┬────┘            │
                                    ┌──────────┼──────────┐       │
                                    │          │          │       │
                              ┌─────▼──┐ ┌────▼────┐ ┌───▼───┐   │
                              │  Win   │ │  Pair   │ │ Draw  │   │
                              │和牌查验│ │碰牌/招/泛│ │起牌出牌│   │
                              └────┬───┘ └────┬────┘ └───┬───┘   │
                                   │          │          │       │
                                   │          └──────────┼───────┘
                                   │                     │
                              ┌────▼─────┐        ┌─────▼──────┐
                              │ Finished │        │ HuangZhuang│
                              │  结束    │        │   黄庄     │
                              └──────────┘        └────────────┘
```

### 5.2 打牌阶段状态子机

```
                 ┌──────────────────────────────────────┐
                 │           Play Phase                  │
                 │                                      │
   CurrentPlayer ├──► Draw from pile ──► Check Win?     │
                 │        │               │   │         │
                 │        │              Yes  No         │
                 │        │               │   │         │
                 │        │           [Win]  ├──►Tong?   │
                 │        │               │   │   │     │
                 │        │               │  Yes  No    │
                 │        │               │   │   │     │
                 │        │               │  [Tong]│    │
                 │        │               │       │     │
                 │        │               │   Discard ──┤
                 │        │               │       │     │
                 │        │               │  ┌────▼─┐   │
                 │        │               │  │Next  │   │
                 │        │               │  │Player│   │
                 │        │               │  │Check │   │
                 │        │               │  │      │   │
                 │        │               │  │Win? ─┼───┤
                 │        │               │  │Pair? │   │
                 │        │               │  │Pass? │   │
                 │        │               │  └──────┘   │
                 │        │                            │
                 │   [Ganta Check]                     │
                 │   [HaiDiLaoYue Check]               │
                 └──────────────────────────────────────┘
```

### 5.3 关键状态转换逻辑

```go
// StateMachine 游戏状态机
type StateMachine struct {
    game   *GameState
    events chan GameEvent
}

// GameEvent 游戏事件
type GameEvent struct {
    Type     EventType
    PlayerID string
    Data     interface{}
}

// HandleEvent 处理游戏事件, 推进状态
func (sm *StateMachine) HandleEvent(evt GameEvent) error {
    switch sm.game.Phase {
    case PhaseCut:
        return sm.handleCut(evt)
    case PhaseDeal:
        return sm.handleDeal(evt)
    case PhaseTongAsk:
        return sm.handleTongAsk(evt)
    case PhasePlay:
        return sm.handlePlay(evt)
    }
    return nil
}

func (sm *StateMachine) handlePlay(evt GameEvent) error {
    switch evt.Type {
    case EventDraw:
        // 起牌 → 检查天胡/自摸 → 可选统牌/赶塔 → 出牌
        tile := sm.game.drawTile(evt.PlayerID)
        if sm.checkSelfWin(evt.PlayerID) {
            sm.transitionWin(evt.PlayerID, WinTypeZiMo)
            return nil
        }
        if sm.canTong(evt.PlayerID) || sm.canGanta(evt.PlayerID) {
            sm.waitForAction(evt.PlayerID) // 等待玩家选择统/赶塔/出牌
        } else {
            sm.waitForDiscard(evt.PlayerID) // 等待出牌
        }

    case EventDiscard:
        // 出牌 → 按优先级询问其他玩家
        sm.game.LastDiscard = evt.Data.(*Tile)
        sm.checkOtherPlayers() // 依次检查闲家是否和牌/碰牌

    case EventPair:
        // 碰牌 → 对/招/泛 → 出牌
        sm.handlePair(evt)

    case EventPass:
        // 过 → 下一个玩家
        sm.advanceTurn()
    }
    return nil
}
```

---

## 6. API 设计

### 6.1 REST API

#### 房间管理

| Method | Path | 说明 |
|--------|------|------|
| POST | /api/v1/rooms | 创建房间 |
| GET | /api/v1/rooms/:id | 查询房间 |
| POST | /api/v1/rooms/:id/join | 加入房间 |
| POST | /api/v1/rooms/:id/leave | 离开房间 |
| POST | /api/v1/rooms/:id/start | 开始游戏(4人齐后) |

#### 游戏操作

| Method | Path | 说明 |
|--------|------|------|
| POST | /api/v1/games/:id/cut | 切牌(歇家/闲二) |
| POST | /api/v1/games/:id/tong | 统牌抉择 |
| POST | /api/v1/games/:id/draw | 起牌 |
| POST | /api/v1/games/:id/discard | 出牌 |
| POST | /api/v1/games/:id/pair | 碰牌(对/招/泛) |
| POST | /api/v1/games/:id/ganta | 赶塔 |
| POST | /api/v1/games/:id/win | 和牌 |
| POST | /api/v1/games/:id/pass | 过 |
| POST | /api/v1/games/:id/dang-jing | 选择当经 |

### 6.2 WebSocket 消息协议

```go
// WSMessage WebSocket消息格式
type WSMessage struct {
    Type    string      `json:"type"`    // 消息类型
    GameID  string      `json:"game_id"` // 牌局ID
    Player  string      `json:"player"`  // 玩家ID
    Data    interface{} `json:"data"`    // 消息体
    Seq     int64       `json:"seq"`     // 序号(用于消息确认)
}
```

#### 服务端推送消息类型

| Type | 说明 | 推送对象 |
|------|------|----------|
| game:start | 牌局开始 | 全体 |
| game:cut-wait | 等待切牌 | 歇家/闲二 |
| game:dealt | 起牌完成(通知手牌) | 各玩家单独 |
| game:tong-ask | 请统阶段 | 全体 |
| game:tong-turn | 轮到某人统牌抉择 | 全体 |
| game:turn | 轮到某人操作 | 全体 |
| game:draw | 某人起牌(不公开牌面) | 全体 |
| game:discard | 某人出牌 | 全体 |
| game:pair | 某人碰牌 | 全体 |
| game:ganta | 某人赶塔 | 全体 |
| game:win | 有人和牌 | 全体 |
| game:huang | 黄庄 | 全体 |
| game:check | 和牌查验(展示手牌) | 全体 |
| player:hand | 手牌更新(私密) | 对应玩家 |

### 6.3 视角隔离

服务端推送时需按玩家视角进行数据过滤：

- **手牌**：仅推送给牌主
- **明牌区**：推送给全体
- **公牌**：不公开内容，仅公开剩余数量
- **出牌区**：推送给全体

---

## 7. AI 玩家设计

### 7.1 决策框架

```go
// AIPlayer AI玩家
type AIPlayer struct {
    player      *Player
    difficulty  AIDifficulty
    memory      *AIMemory // 记忆其他玩家出牌
}

type AIDifficulty int

const (
    AIEasy   AIDifficulty = iota // 简单: 随机决策
    AIMedium                     // 中等: 基本策略
    AIHard                       // 困难: 完整策略
)

// AIDecision AI决策结果
type AIDecision struct {
    Action    TurnAction
    Tile      *Tile     // 出牌/碰牌对应的牌
    TongSize  int       // 统牌大小(4或5)
    DangJing  TileName  // 当经选择
}
```

### 7.2 决策策略

基于 `strategy.md` 的策略，按优先级排列：

```go
func (ai *AIPlayer) Decide(ctx *GameContext) AIDecision {
    // 1. 检查是否可以和牌
    if ai.canWin(ctx) {
        return AIDecision{Action: ActionWin}
    }

    // 2. 选择当经(首次)
    if ai.needChooseDangJing() {
        jing := ai.chooseDangJing(ctx)
        return AIDecision{Action: ActionDangJing, DangJing: jing}
    }

    // 3. 是否碰牌
    if ai.canPair(ctx) {
        if ai.shouldPair(ctx) {
            return ai.decidePairAction(ctx)
        }
        return AIDecision{Action: ActionPass}
    }

    // 4. 是否统牌
    if ai.canTong(ctx) && ai.shouldTong(ctx) {
        return ai.decideTongAction(ctx)
    }

    // 5. 是否赶塔
    if ai.canGanta(ctx) {
        return AIDecision{Action: ActionGanta}
    }

    // 6. 出牌决策
    return ai.decideDiscard(ctx)
}
```

### 7.3 出牌策略

```go
func (ai *AIPlayer) decideDiscard(ctx *GameContext) AIDecision {
    candidates := ai.evaluateDiscardCandidates(ctx)

    // 按策略评分排序
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Score < candidates[j].Score // 分越低越优先打出
    })

    // 跟张策略: 如果判断需防守, 优先打出别人打过的牌
    if ai.shouldPlaySafe(ctx) {
        if safe := ai.findSafeDiscard(ctx, candidates); safe != nil {
            return AIDecision{Action: ActionDiscard, Tile: safe}
        }
    }

    return AIDecision{Action: ActionDiscard, Tile: candidates[0].Tile}
}

// evaluateDiscardCandidates 评估每张牌的出牌价值
// 评分维度:
// 1. 该牌是否可参与组合(不可参与的优先打)
// 2. 该牌打出后的胡数变化(胡数降低少的优先打)
// 3. 该牌打出后的听牌变化(影响听牌的延后打)
// 4. 该牌是否为经牌(经牌尽量不打)
// 5. 别字绝对不打
func (ai *AIPlayer) evaluateDiscardCandidates(ctx *GameContext) []DiscardCandidate {
    var candidates []DiscardCandidate
    hand := ai.player.Hand

    for _, tile := range hand {
        if tile.Name == TileBie {
            continue // 别字绝对不打
        }

        score := 0.0

        // 组合参与度
        combCount := ai.countCombinationsInvolving(tile, ctx)
        score += float64(combCount) * 10

        // 胡数贡献
        huBefore := ai.totalHu(ctx)
        huAfter := ai.totalHuWithout(tile, ctx)
        score += float64(huBefore-huAfter) * 5

        // 经牌权重
        if tile.IsJing() {
            score += 20
            if tile.IsFlower {
                score += 15 // 花经更不能打
            }
        }

        // 红色牌权重(一般比黑色更有价值)
        if tile.Color == ColorRed {
            score += 3
        }

        candidates = append(candidates, DiscardCandidate{Tile: tile, Score: score})
    }

    return candidates
}
```

---

## 8. 目录结构

```
f3moon/
├── cmd/
│   └── server/
│       └── main.go              # 服务端入口
├── internal/
│   ├── model/
│   │   ├── tile.go              # 牌定义与初始化
│   │   ├── combination.go       # 组合定义
│   │   ├── player.go            # 玩家模型
│   │   └── game.go              # 牌局状态模型
│   ├── engine/
│   │   ├── deck.go              # 洗牌与牌堆管理
│   │   ├── hu_calc.go           # 胡数计算
│   │   ├── hu_table.go          # 胡数查表
│   │   ├── arrange.go           # 理牌算法(回溯搜索)
│   │   ├── ting.go              # 听牌检测
│   │   ├── dangjing.go          # 当经选择
│   │   ├── bie_resolver.go      # 别字多义处理
│   │   └── validator.go         # 和牌校验
│   ├── game/
│   │   ├── state_machine.go     # 游戏状态机
│   │   ├── phase_cut.go         # 切牌阶段
│   │   ├── phase_deal.go        # 起牌阶段
│   │   ├── phase_tong.go        # 请统阶段
│   │   ├── phase_play.go        # 打牌阶段
│   │   ├── phase_check.go       # 查验阶段
│   │   └── event.go             # 事件定义
│   ├── ai/
│   │   ├── player.go            # AI玩家主体
│   │   ├── strategy.go          # 策略决策
│   │   ├── discard.go           # 出牌策略
│   │   ├── pair.go              # 碰牌策略
│   │   ├── tong.go              # 统牌策略
│   │   └── memory.go            # 出牌记忆
│   ├── room/
│   │   ├── room.go              # 房间管理
│   │   └── manager.go           # 房间管理器
│   ├── ws/
│   │   ├── hub.go               # WebSocket连接管理
│   │   ├── client.go            # WebSocket客户端
│   │   └── message.go           # 消息定义
│   ├── handler/
│   │   ├── room.go              # 房间HTTP接口
│   │   ├── game.go              # 游戏HTTP接口
│   │   └── ws.go                # WebSocket接口
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   └── util/
│       └── rand.go              # 随机工具
├── design/
│   └── server.md                # 本设计文档
├── rules.md                     # 游戏规则
├── strategy.md                  # 博弈策略
├── go.mod
└── go.sum
```

---

## 9. 关键设计决策

### 9.1 别字（"别"）的处理

"别"字是本系统最大的设计挑战。核心方案：

- **延迟绑定**：别字不预先指定视作哪个字，而是在理牌/算胡时动态确定
- **枚举搜索**：2张别字最多有 3^2 = 9 种分配方案（每张可作三/五/七），对每种方案执行理牌算法
- **最优选择**：在所有方案中选取胡数最大的方案作为最终结果
- **约束传播**：如果别字参与文字组合（如"化别千"），则其视作范围缩小，可加速搜索

### 9.2 理牌算法选择

| 方案 | 优点 | 缺点 |
|------|------|------|
| 回溯搜索 | 结果完备，能找到所有合法组合 | 搜索空间大，可能慢 |
| 贪心算法 | 速度快 | 可能错过最优解 |
| 贪心+局部搜索 | 兼顾速度和质量 | 实现复杂 |

**选择**：采用回溯搜索 + 剪枝优化。理由：
- 手牌最多26张，搜索空间可控
- 必须找到所有听牌方案，贪心无法保证
- 剪枝后性能可接受

### 9.3 并发模型

```go
// 每个牌局一个 goroutine, 串行处理事件
// 避免并发问题, 简化状态管理
func (g *GameManager) RunGame(gameID string) {
    sm := NewStateMachine(gameID)
    for evt := range sm.events {
        sm.HandleEvent(evt)
        sm.broadcastState() // 推送状态给所有玩家
    }
}
```

- 每个牌局运行在独立 goroutine 中，事件通过 channel 串行处理
- WebSocket Hub 独立运行，负责消息广播
- 房间管理器使用读写锁保护共享状态

### 9.4 超时与断线处理

| 场景 | 处理方式 |
|------|----------|
| 玩家操作超时(30s) | 自动选择"过"或最优决策(按简单AI策略) |
| 玩家断线 | 保留座位60s，超时后由AI接管 |
| AI玩家 | 无超时问题，立即响应 |
| 全员断线 | 牌局暂停，5分钟后自动解散 |

---

## 10. 测试策略

### 10.1 单元测试重点

| 模块 | 测试要点 |
|------|----------|
| tile.go | 112张牌的生成与属性正确性 |
| hu_calc.go | 所有胡数规则的覆盖(按规则文档逐条验证) |
| hu_table.go | 经牌组合查表的完整性 |
| arrange.go | 各种手牌组合的理牌正确性 |
| ting.go | 听牌检测与听字计算 |
| bie_resolver.go | 别字的所有分配方案与边界情况 |
| state_machine.go | 状态流转的完整性 |
| validator.go | 和牌校验的准确性 |

### 10.2 集成测试

- 完整牌局流程的端到端测试
- 多玩家WebSocket连接的并发测试
- AI玩家的对局模拟测试

### 10.3 规则合规性测试

基于 rules.md 中的每个示例编写测试用例：

```go
func TestHuCalc_Examples(t *testing.T) {
    // 规则第177行: "孔乙己"一句，"乙" 不带花，算作 0 胡
    comb := makeWordComb("孔", "乙", "己") // 乙不带花
    assert.Equal(t, 0, CalcHu(comb, TileWu))

    // 规则第180行: "孔乙己"一句，"乙" 带花，算作 1 胡
    comb = makeWordComb("孔", "花乙", "己")
    assert.Equal(t, 1, CalcHu(comb, TileWu))

    // 规则第250行: 3张白"七"，当经算为10胡
    comb = makeSingleComb(3, "七", flower: 0)
    assert.Equal(t, 10, CalcHu(comb, TileQi))
}
```

---

## 11. 待确认事项

以下问题在 rules.md 中未明确，需要在开发前确认：

1. **黄庄后庄家轮换**：黄庄后，下一局庄家是否轮换？如何轮换？
2. **和牌后庄家轮换**：和牌后，下一局庄家是继续坐庄还是轮换？
3. **积分结算规则**：输家付分与爬坡的换算关系（每坡多少分/多少倍）？
4. **请统后庄家手牌**：庄家请统时起第26张牌，该牌是否参与请统判断？
5. **"别"参与碰牌**：规则说碰牌"不包含别字"，但"别"可参与单字组合，该组合能否放入明牌区？
6. **数字组合中的"别"**：规则列举了别字替换三/五/七的数字组合，但未明确别字是否可以替换数字组合中的非经数字位（如别是否可替换"乙"作1）——从规则推论应不可，需确认。
7. **黑色五张统带花**：规则第246行提到"黑色带双花五张统6胡"，但实际5张牌中必含2花，是否存在"带1花五张统"的情况？理论上5张同字牌中，乙/三/五/七/九有2花，其余无花，所以黑色五张统要么无花要么双花，不存在单花情况，需确认。
