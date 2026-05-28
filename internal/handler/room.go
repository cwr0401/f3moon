package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/model"
	"github.com/cwr0401/f3moon/internal/room"
)

// RoomHandler 房间接口处理器
type RoomHandler struct {
	manager *room.Manager
}

// NewRoomHandler 创建房间处理器
func NewRoomHandler(manager *room.Manager) *RoomHandler {
	return &RoomHandler{manager: manager}
}

// CreateRoom 创建房间
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req struct {
		Name     string         `json:"name" binding:"required"`
		Mode     model.GameMode `json:"mode"`
		PlayerID string         `json:"player_id" binding:"required"`
		PlayerName string       `json:"player_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r := h.manager.CreateRoom(req.Name, req.Mode, req.PlayerID)
	player := &room.RoomPlayer{
		ID:   req.PlayerID,
		Name: req.PlayerName,
	}
	r.AddPlayer(player)

	c.JSON(http.StatusOK, gin.H{
		"room_id": r.ID,
		"room":    r,
	})
}

// GetRoom 查询房间
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
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms := h.manager.ListRooms()
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// JoinRoom 加入房间
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	var req struct {
		PlayerID   string `json:"player_id" binding:"required"`
		PlayerName string `json:"player_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := &room.RoomPlayer{
		ID:   req.PlayerID,
		Name: req.PlayerName,
	}
	r, err := h.manager.JoinRoom(roomID, player)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

// LeaveRoom 离开房间
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.LeaveRoom(roomID, req.PlayerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// AddAIPlayer 添加AI玩家
func (h *RoomHandler) AddAIPlayer(c *gin.Context) {
	roomID := c.Param("id")
	r := h.manager.GetRoom(roomID)
	if r == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	player := r.AddAIPlayer()
	if player == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room is full"})
		return
	}
	c.JSON(http.StatusOK, player)
}

// StartGame 开始游戏
func (h *RoomHandler) StartGame(c *gin.Context) {
	roomID := c.Param("id")
	r := h.manager.GetRoom(roomID)
	if r == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if !r.IsFull() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room is not full"})
		return
	}
	if !r.AllReady() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not all players ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "started", "room_id": roomID})
}
