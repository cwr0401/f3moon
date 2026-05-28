package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/game"
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
func (h *GameHandler) Cut(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
		Position int    `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventCut,
		PlayerID: req.PlayerID,
		Data:     game.CutData{Position: req.Position},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Tong 统牌抉择
func (h *GameHandler) Tong(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string          `json:"player_id" binding:"required"`
		TileName model.TileName  `json:"tile_name"`
		TongSize int             `json:"tong_size"`
		Skip     bool            `json:"skip"` // 不统牌
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
		PlayerID: req.PlayerID,
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
func (h *GameHandler) Draw(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventDraw,
		PlayerID: req.PlayerID,
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Discard 出牌
func (h *GameHandler) Discard(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
		TileID   int    `json:"tile_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventDiscard,
		PlayerID: req.PlayerID,
		Data:     game.DiscardData{TileID: req.TileID},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Pair 碰牌
func (h *GameHandler) Pair(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
		TileID   int    `json:"tile_id" binding:"required"`
		PairSize int    `json:"pair_size" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventPair,
		PlayerID: req.PlayerID,
		Data:     game.PairData{TileID: req.TileID, PairSize: req.PairSize},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ganta 赶塔
func (h *GameHandler) Ganta(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventGanta,
		PlayerID: req.PlayerID,
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Win 和牌
func (h *GameHandler) Win(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventWin,
		PlayerID: req.PlayerID,
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Pass 过
func (h *GameHandler) Pass(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventPass,
		PlayerID: req.PlayerID,
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DangJing 选择当经
func (h *GameHandler) DangJing(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	var req struct {
		PlayerID string          `json:"player_id" binding:"required"`
		Jing     model.TileName  `json:"jing" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evt := game.GameEvent{
		Type:     game.EventDangJing,
		PlayerID: req.PlayerID,
		Data:     game.DangJingData{Jing: req.Jing},
	}

	if err := sm.HandleEvent(evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetGame 查询游戏状态
func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	sm, ok := h.games[gameID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, sm.Game())
}
