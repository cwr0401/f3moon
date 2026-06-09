package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/middleware"
	"github.com/cwr0401/f3moon/internal/model"
)

// GameHandler 游戏接口处理器
type GameHandler struct {
	games map[string]*game.StateMachine
}

// NewGameHandler 创建游戏处理器
func NewGameHandler() *GameHandler {
	return &GameHandler{
		games: make(map[string]*game.StateMachine),
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

	evt := game.GameEvent{
		Type:     game.EventCut,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
		Data:     game.CutData{Position: req.Position},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

	evtType := game.EventTong
	if req.Skip {
		evtType = game.EventPass
	}

	evt := game.GameEvent{
		Type:     evtType,
		PlayerID: c.GetString(middleware.ContextKeyUserID),
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
