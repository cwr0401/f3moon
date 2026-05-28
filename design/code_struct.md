# 项目结构

```
  f3moon/
  ├── cmd/server/main.go              # 服务端入口, Gin路由注册
  ├── internal/
  │   ├── model/
  │   │   ├── tile.go                 # 牌定义(112张)、颜色、经牌等
  │   │   ├── combination.go          # 组合定义(文字/数字/单字)、别字映射
  │   │   ├── player.go               # 玩家模型(手牌、明牌、当经等)
  │   │   └── game.go                 # 牌局状态(阶段、公牌、出牌区等)
  │   ├── engine/
  │   │   ├── deck.go                 # 洗牌(Fisher-Yates)、切牌、起牌
  │   │   ├── hu_table.go             # 经牌组合胡数查表(27条规则)
  │   │   ├── hu_calc.go              # 胡数计算(分层公式:基础+花+经)
  │   │   ├── bie_resolver.go         # 别字多义处理(枚举3^n方案)
  │   │   ├── arrange.go              # 理牌算法(回溯搜索+剪枝)
  │   │   ├── ting.go                 # 听牌检测与听字计算
  │   │   ├── dangjing.go             # 当经选择(枚举三/五/七取最优)
  │   │   └── validator.go            # 和牌校验、碰牌/赶塔判断
  │   ├── game/
  │   │   ├── event.go                # 事件定义、推送消息类型
  │   │   ├── state_machine.go        # 游戏状态机核心
  │   │   ├── phase_cut.go            # 切牌阶段
  │   │   ├── phase_deal.go           # 起牌+当经选择+进入请统
  │   │   ├── phase_tong.go           # 请统阶段+进入打牌
  │   │   ├── phase_play.go           # 打牌阶段(起牌/出牌/碰牌/赶塔/和牌/海底捞月)
  │   │   └── phase_check.go          # 和牌查验
  │   ├── ai/
  │   │   ├── player.go               # AI玩家主体+决策框架
  │   │   ├── memory.go               # 出牌记忆(跟张策略依据)
  │   │   ├── strategy.go             # 碰牌/赶塔/防守策略(基于strategy.md)
  │   │   ├── discard.go              # 出牌策略(三级评分)
  │   │   ├── pair.go                 # 碰牌决策
  │   │   └── tong.go                 # 统牌决策+模拟胡数
  │   ├── room/
  │   │   ├── room.go                 # 房间管理(加入/离开/就绪/AI填充)
  │   │   └── manager.go              # 房间管理器
  │   ├── ws/
  │   │   ├── hub.go                  # WebSocket连接管理(广播/定向推送)
  │   │   ├── client.go               # WebSocket客户端(读写泵)
  │   │   └── message.go              # 消息格式定义
  │   ├── handler/
  │   │   ├── room.go                 # 房间HTTP接口
  │   │   ├── game.go                 # 游戏HTTP接口
  │   │   └── ws.go                   # WebSocket接口
  │   └── config/
  │       └── config.go               # 配置管理(环境变量)
  ├── pkg/util/rand.go                # 随机数工具
  └── bin/f3moon                      # 编译产物(17MB)
```

## 关键设计要点

- 别字处理：枚举 3^n 种分配方案（每张别可作花三/花五/花七），对每种方案执行回溯理牌
- 胡数计算：非经牌用"基础胡+花加成"公式，经牌组合查表（27条规则全覆盖）
- AI策略：三级难度，困难模式包含跟张策略、撬张策略、胡数-听牌平衡评估
- 编译验证：go build + go vet 均通过

claude --resume e455361a-ce5a-4a79-8a7b-4219efe7ed7c
