package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/middleware"
	"github.com/cwr0401/f3moon/internal/model"
)

// GameHandler 游戏接口处理器
type GameHandler struct {
	games    map[string]*game.StateMachine
	gameRepo game.Repository
}

// NewGameHandler 创建游戏处理器
func NewGameHandler(gameRepo game.Repository) *GameHandler {
	return &GameHandler{
		games:    make(map[string]*game.StateMachine),
		gameRepo: gameRepo,
	}
}

// RegisterGame 注册游戏实例
func (h *GameHandler) RegisterGame(gameID string, sm *game.StateMachine) {
	h.games[gameID] = sm
}

// Cut 切牌
// @Summary 切牌（腰牌）
// @Description 歇家或闲二选择切牌位置（37-110）
// @Tags 游戏
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Param body body object true "切牌位置" example({"position":73})
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/cut [post]
func (h *GameHandler) Cut(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		Position int `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 从数据库获取游戏牌栈记录
	deck, err := h.gameRepo.GetGameDeckByRoomID(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "game deck not found"})
		return
	}

	// 2. 验证是否已完成洗牌
	if !deck.Shuffled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game not shuffled yet"})
		return
	}

	// 3. 验证是否已完成切牌
	if deck.CutFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already cut"})
		return
	}

	// 4. 验证切牌位置范围
	position := req.Position
	if position < 37 || position > 110 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cut position out of range [37, 110]"})
		return
	}

	// 5. 执行切牌：将 0~position 的牌从栈顶弹出，从栈底压入
	cutStack := engine.CutIDs(deck.ShuffleStack, position)

	// 6. 更新数据库记录
	deck.CutStack = cutStack
	deck.CutPosition = position
	deck.CutFinished = true
	deck.UpdatedAt = time.Now().UTC()

	if err := h.gameRepo.UpdateGameDeck(deck); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game deck"})
		return
	}

	// 7. 同时更新内存中的游戏状态（转换ID为Tile对象）
	gs := sm.Game()
	gs.DrawPile = model.GetTilesFromIDs(cutStack)
	gs.Phase = model.PhaseDeal // 更新Phase到发牌阶段

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Tong 统牌抉择
// @Summary 统牌抉择
// @Description 请统阶段，玩家选择统牌或跳过
// @Tags 游戏
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Param body body object true "统牌信息" example({"tile_name":"三","tong_size":4,"skip":false})
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/tong [post]
func (h *GameHandler) Tong(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		TileName model.TileName `json:"tile_name"`
		TongSize int            `json:"tong_size"`
		Skip     bool           `json:"skip"` // 不统牌
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 从数据库获取游戏牌栈记录
	deck, err := h.gameRepo.GetGameDeckByRoomID(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "game deck not found"})
		return
	}

	// 2. 验证是否完成洗牌
	if !deck.Shuffled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game not shuffled yet"})
		return
	}

	// 3. 验证是否完成切牌
	if !deck.CutFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game not cut yet"})
		return
	}

	// 4. 验证是否完成发牌
	if !deck.DealFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game not dealt yet"})
		return
	}

	// 5. 验证是否已完成请统
	if deck.TongFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tong already finished"})
		return
	}

	playerID := c.GetString(middleware.ContextKeyUserID)
	gs := sm.Game()

	// 6. 验证是否是当前玩家的回合
	tongOrder := []int{2, 1, 0} // 闲二 -> 闲一 -> 庄家
	if len(deck.TongOrder) == 3 {
		tongOrder = []int{int(deck.TongOrder[0]), int(deck.TongOrder[1]), int(deck.TongOrder[2])}
	}
	currentIdx := tongOrder[deck.TongCurrent]
	currentPlayer := gs.Players[currentIdx]
	if currentPlayer == nil || currentPlayer.ID != playerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not your turn"})
		return
	}

	// 7. 如果不跳过，验证手牌并操作数据库牌栈
	if !req.Skip {
		// 验证统牌大小
		if req.TongSize != 4 && req.TongSize != 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tong_size must be 4 or 5"})
			return
		}

		// 验证牌名
		if req.TileName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tile_name is required"})
			return
		}

		// 获取当前玩家的手牌ID数组
		var playerHand []uint8
		switch currentIdx {
		case 0:
			playerHand = deck.DealerHand
		case 1:
			playerHand = deck.Player1Hand
		case 2:
			playerHand = deck.Player2Hand
		}

		// 统计手牌中该牌的数量并构建新手牌（移除统掉的牌）
		count := 0
		var newHand []uint8
		for _, tileID := range playerHand {
			tile := model.GetTile(tileID)
			if tile != nil && tile.Name == req.TileName && count < req.TongSize {
				count++
				continue // 跳过，不移到新手牌中
			}
			newHand = append(newHand, tileID)
		}

		if count < req.TongSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not enough tiles to tong"})
			return
		}

		// 计算需要补充的牌数（统4张补1张，统5张补2张）
		drawCount := req.TongSize - 3
		cutStack := deck.CutStack
		stackBottom := deck.StackBottom

		// 验证栈底有足够的牌
		availableCards := stackBottom - deck.StackTop + 1
		if availableCards < drawCount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not enough cards in draw pile"})
			return
		}

		// 从栈底取出牌
		var drawnTileIDs []uint8
		for i := 0; i < drawCount; i++ {
			drawnTileIDs = append(drawnTileIDs, cutStack[stackBottom])
			stackBottom--
		}

		// 将取出的牌加入玩家手牌
		newHand = append(newHand, drawnTileIDs...)

		// 更新数据库中的手牌
		switch currentIdx {
		case 0:
			deck.DealerHand = newHand
		case 1:
			deck.Player1Hand = newHand
		case 2:
			deck.Player2Hand = newHand
		}
		deck.StackBottom = stackBottom
	}

	// 8. 更新数据库中的请统状态
	deck.TongCurrent = deck.TongCurrent + 1
	if deck.TongCurrent >= 3 {
		deck.TongFinished = true
	}
	// 只有在TongOrder未设置时才设置默认值
	if len(deck.TongOrder) == 0 {
		deck.TongOrder = game.Uint8Slice{2, 1, 0}
	}

	// 同步内存中的游戏状态（从数据库更新）
	if !req.Skip {
		gs.Players[0].Hand = model.GetTilesFromIDs(deck.DealerHand)
		gs.Players[1].Hand = model.GetTilesFromIDs(deck.Player1Hand)
		gs.Players[2].Hand = model.GetTilesFromIDs(deck.Player2Hand)
		gs.DrawPile = model.GetTilesFromIDs(deck.CutStack[deck.StackTop : deck.StackBottom+1])
	} else {
		// 跳过统牌时也需要确保内存状态与数据库一致
		gs.Players[0].Hand = model.GetTilesFromIDs(deck.DealerHand)
		gs.Players[1].Hand = model.GetTilesFromIDs(deck.Player1Hand)
		gs.Players[2].Hand = model.GetTilesFromIDs(deck.Player2Hand)
	}

	deck.UpdatedAt = time.Now().UTC()

	if err := h.gameRepo.UpdateGameDeck(deck); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game deck"})
		return
	}

	// 9. 处理事件（持久化后再处理）
	evtType := game.EventTong
	if req.Skip {
		evtType = game.EventPass
	}

	evt := game.GameEvent{
		Type:     evtType,
		PlayerID: playerID,
		Data: game.TongData{
			TileName: req.TileName,
			TongSize: req.TongSize,
		},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Draw 起牌
// @Summary 起牌
// @Description 从公牌顶部起一张牌
// @Tags 游戏
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/draw [post]
func (h *GameHandler) Draw(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventDraw,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Discard 出牌
// @Summary 出牌
// @Description 打出一张手牌
// @Tags 游戏
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Param body body object true "出牌信息" example({"tile_id":42})
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/discard [post]
func (h *GameHandler) Discard(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		TileID int `json:"tile_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventDiscard,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
		Data:     game.DiscardData{TileID: req.TileID},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Pair 碰牌
// @Summary 碰牌（对/招/泛）
// @Description 对别人打出的牌进行碰牌操作
// @Tags 游戏
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Param body body object true "碰牌信息" example({"tile_id":42,"pair_size":3})
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/pair [post]
func (h *GameHandler) Pair(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		TileID   int `json:"tile_id" binding:"required"`
		PairSize int `json:"pair_size" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventPair,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
		Data:     game.PairData{TileID: req.TileID, PairSize: req.PairSize},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ganta 赶塔
// @Summary 赶塔
// @Description 将手中同名牌加入已有的统组合中
// @Tags 游戏
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/ganta [post]
func (h *GameHandler) Ganta(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventGanta,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Win 和牌
// @Summary 和牌（胡牌）
// @Description 声明和牌，系统验证手牌是否满足和牌条件
// @Tags 游戏
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/win [post]
func (h *GameHandler) Win(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventWin,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Pass 过
// @Summary 过
// @Description 放弃当前操作（碰牌、和牌等）
// @Tags 游戏
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/pass [post]
func (h *GameHandler) Pass(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventPass,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DangJing 选择当经
// @Summary 选择当经
// @Description 选择当经牌（三、五、七之一）
// @Tags 游戏
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Param body body object true "当经选择" example({"jing":"三"})
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/dang-jing [post]
func (h *GameHandler) DangJing(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		Jing model.TileName `json:"jing" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventDangJing,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
		Data:     game.DangJingData{Jing: req.Jing},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Deal 发牌
// @Summary 发牌
// @Description 执行发牌流程，给三个玩家发牌
// @Tags 游戏
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id}/deal [post]
func (h *GameHandler) Deal(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	// 1. 从数据库获取游戏牌栈记录
	deck, err := h.gameRepo.GetGameDeckByRoomID(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "game deck not found"})
		return
	}

	// 2. 验证是否完成洗牌
	if !deck.Shuffled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game not shuffled yet"})
		return
	}

	// 3. 验证是否完成切牌
	if !deck.CutFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game not cut yet"})
		return
	}

	// 4. 验证是否已完成发牌
	if deck.DealFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already dealt"})
		return
	}

	// 5. 从切牌栈数据构建栈
	cutStack := deck.CutStack
	stackTop := deck.StackTop

	// 6. 依次从栈顶弹出牌ID，按庄家、闲家一、闲家二的顺序发牌，每人25张
	var dealerHand, player1Hand, player2Hand []uint8
	for i := 0; i < 25; i++ {
		// 庄家
		dealerHand = append(dealerHand, cutStack[stackTop])
		stackTop++
		// 闲家一
		player1Hand = append(player1Hand, cutStack[stackTop])
		stackTop++
		// 闲家二
		player2Hand = append(player2Hand, cutStack[stackTop])
		stackTop++
	}

	// 7. 给庄家再发1张牌
	dealerHand = append(dealerHand, cutStack[stackTop])
	stackTop++

	// 8. 从游戏状态获取玩家ID
	gs := sm.Game()
	var dealerID, player1ID, player2ID string
	if len(gs.Players) >= 3 {
		dealerID = gs.Players[0].ID
		player1ID = gs.Players[1].ID
		player2ID = gs.Players[2].ID
	}

	// 9. 更新数据库记录
	deck.StackTop = stackTop
	deck.StackBottom = len(cutStack) - 1
	deck.DealerID = dealerID
	deck.Player1ID = player1ID
	deck.Player2ID = player2ID
	deck.DealerHand = dealerHand
	deck.Player1Hand = player1Hand
	deck.Player2Hand = player2Hand
	deck.DealerFinished = true
	deck.DealFinished = true
	deck.UpdatedAt = time.Now().UTC()

	if err := h.gameRepo.UpdateGameDeck(deck); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game deck"})
		return
	}

	// 10. 同时更新内存中的游戏状态
	gs.Players[0].Hand = model.GetTilesFromIDs(dealerHand)
	gs.Players[1].Hand = model.GetTilesFromIDs(player1Hand)
	gs.Players[2].Hand = model.GetTilesFromIDs(player2Hand)
	gs.DrawPile = model.GetTilesFromIDs(cutStack[stackTop:])

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetGame 查询游戏状态
// @Summary 查询游戏状态
// @Description 获取指定游戏的完整状态
// @Tags 游戏
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏 ID"
// @Success 200 {object} object "游戏状态"
// @Failure 404 {object} object "{ \"error\": \"game not found\" }"
// @Router /games/{id} [get]
func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, sm.Game())
}
