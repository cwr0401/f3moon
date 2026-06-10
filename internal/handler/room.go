package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/middleware"
	"github.com/cwr0401/f3moon/internal/model"
	"github.com/cwr0401/f3moon/internal/room"
	"github.com/cwr0401/f3moon/internal/ws"
)

// RoomHandler 房间接口处理器
type RoomHandler struct {
	manager      *room.Manager
	gameHandler  *GameHandler
	hub          *ws.Hub
	gameRepo     game.Repository
}

// NewRoomHandler 创建房间处理器
func NewRoomHandler(manager *room.Manager, gameHandler *GameHandler, hub *ws.Hub, gameRepo game.Repository) *RoomHandler {
	return &RoomHandler{manager: manager, gameHandler: gameHandler, hub: hub, gameRepo: gameRepo}
}

// CreateRoom 创建房间
// @Summary 创建房间
// @Description 创建新游戏房间，创建者自动加入
// @Tags 房间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "房间信息" example({"name":"花牌局","mode":0})
// @Success 200 {object} object "{ \"room_id\": \"string\", \"room\": {} }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Router /rooms [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	var req struct {
		Name string         `json:"name" binding:"required"`
		Mode model.GameMode `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r, err := h.manager.CreateRoom(req.Name, req.Mode, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	player := &room.RoomPlayer{
		ID:   userID,
		Name: c.GetString(middleware.ContextKeyNickname),
	}
	if _, err := h.manager.JoinRoom(r.ID, player); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id": r.ID,
		"room":    r,
	})
}

// GetRoom 查询房间
// @Summary 查询房间
// @Description 根据房间 ID 查询房间详情
// @Tags 房间
// @Produce json
// @Security BearerAuth
// @Param id path string true "房间 ID"
// @Success 200 {object} object "房间信息"
// @Failure 404 {object} object "{ \"error\": \"room not found\" }"
// @Router /rooms/{id} [get]
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")
	r := h.manager.GetRoom(roomID)
	if r == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	c.JSON(http.StatusOK, r)
}

// ListRooms 列出所有房间
// @Summary 列出所有房间
// @Description 获取所有游戏房间列表
// @Tags 房间
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "{ \"rooms\": [] }"
// @Router /rooms [get]
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms := h.manager.ListRooms()
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// JoinRoom 加入房间
// @Summary 加入房间
// @Description 加入指定房间
// @Tags 房间
// @Produce json
// @Security BearerAuth
// @Param id path string true "房间 ID"
// @Success 200 {object} object "房间信息"
// @Failure 400 {object} object "{ \"error\": \"room is full\" }"
// @Router /rooms/{id}/join [post]
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	player := &room.RoomPlayer{
		ID:   userID,
		Name: userID,
	}
	r, err := h.manager.JoinRoom(roomID, player)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

// LeaveRoom 离开房间
// @Summary 离开房间
// @Description 离开指定房间
// @Tags 房间
// @Produce json
// @Security BearerAuth
// @Param id path string true "房间 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Router /rooms/{id}/leave [post]
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	if err := h.manager.LeaveRoom(roomID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// AddAIPlayer 添加AI玩家
// @Summary 添加AI玩家
// @Description 向房间添加一个AI玩家（AI自动准备）
// @Tags 房间
// @Produce json
// @Security BearerAuth
// @Param id path string true "房间 ID"
// @Success 200 {object} object "AI 玩家信息"
// @Failure 400 {object} object "{ \"error\": \"room is full\" }"
// @Failure 404 {object} object "{ \"error\": \"room not found\" }"
// @Router /rooms/{id}/ai [post]
func (h *RoomHandler) AddAIPlayer(c *gin.Context) {
	roomID := c.Param("id")
	player, err := h.manager.AddAIPlayer(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, player)
}

// Ready 设置玩家准备状态
// @Summary 设置准备状态
// @Description 设置当前玩家的准备或取消准备状态
// @Tags 房间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "房间 ID"
// @Param body body object true "准备状态" example({"ready":true})
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 404 {object} object "{ \"error\": \"room not found\" }"
// @Router /rooms/{id}/ready [post]
func (h *RoomHandler) Ready(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	var req struct {
		Ready bool `json:"ready"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r, err := h.manager.SetReady(roomID, userID, req.Ready)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = r
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// StartGame 开始游戏
// @Summary 开始游戏
// @Description 房主开始游戏，需要房间已满且所有玩家已准备
// @Tags 房间
// @Produce json
// @Security BearerAuth
// @Param id path string true "房间 ID"
// @Success 200 {object} object "{ \"status\": \"started\", \"room_id\": \"string\" }"
// @Failure 400 {object} object "{ \"error\": \"room is not full\" }"
// @Failure 404 {object} object "{ \"error\": \"room not found\" }"
// @Router /rooms/{id}/start [post]
func (h *RoomHandler) StartGame(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	r, err := h.manager.StartRoomGame(roomID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建并保存游戏牌栈记录
	seed := time.Now().UnixNano()
	shuffleStack := engine.NewShuffledDeckStack(seed)
	deck := &game.GameDeckRecord{
		ID:           uuid.NewString(),
		RoomID:       roomID,
		Seed:         seed,
		ShuffleStack: shuffleStack,
		CutStack:     nil,
		StackTop:     0,
		StackBottom:  111,
		Shuffled:     true,
		CutPosition:  0,
		CutFinished:  false,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := h.gameRepo.CreateGameDeck(deck); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game deck"})
		return
	}

	// 创建游戏状态，并使用数据库中的牌栈
	gs := model.NewGameState(roomID, r.Mode)
	gs.Players = r.ToModelPlayers()
	gs.DrawPile = model.GetTilesFromIDs(shuffleStack)

	// 创建状态机并启动游戏
	sm := game.NewStateMachine(gs, h.hub)
	if err := sm.StartGame(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 注册游戏实例
	h.gameHandler.RegisterGame(roomID, sm)

	c.JSON(http.StatusOK, gin.H{"status": "started", "room_id": roomID})
}
