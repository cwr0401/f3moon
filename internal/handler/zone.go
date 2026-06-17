package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/zone"
)

// ZoneHandler 游戏区接口处理器
type ZoneHandler struct {
	zoneManager *zone.Manager
}

// NewZoneHandler 创建游戏区处理器
func NewZoneHandler(zoneManager *zone.Manager) *ZoneHandler {
	return &ZoneHandler{zoneManager: zoneManager}
}

// CreateZone 创建游戏区
// @Summary 创建游戏区
// @Description 创建新的游戏区
// @Tags 游戏区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "游戏区信息" example({"name":"新手区","description":"适合新手玩家","max_rooms":10})
// @Success 200 {object} object "{ \"zone\": {} }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Router /zones [post]
func (h *ZoneHandler) CreateZone(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		MaxRooms    int    `json:"max_rooms" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	z, err := h.zoneManager.CreateZone(req.Name, req.Description, req.MaxRooms)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zone": z})
}

// ListZones 列出所有游戏区
// @Summary 列出所有游戏区
// @Description 获取所有游戏区列表
// @Tags 游戏区
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "{ \"zones\": [] }"
// @Router /zones [get]
func (h *ZoneHandler) ListZones(c *gin.Context) {
	zones := h.zoneManager.ListZones()
	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

// GetZone 获取游戏区详情
// @Summary 获取游戏区详情
// @Description 根据游戏区 ID 获取详情
// @Tags 游戏区
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏区 ID"
// @Success 200 {object} object "游戏区信息"
// @Failure 404 {object} object "{ \"error\": \"zone not found\" }"
// @Router /zones/{id} [get]
func (h *ZoneHandler) GetZone(c *gin.Context) {
	zoneID := c.Param("id")
	z := h.zoneManager.GetZone(zoneID)
	if z == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
		return
	}
	c.JSON(http.StatusOK, z)
}

// UpdateZone 更新游戏区
// @Summary 更新游戏区
// @Description 更新游戏区信息
// @Tags 游戏区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏区 ID"
// @Param body body object true "游戏区信息" example({"name":"新手区","description":"适合新手玩家","max_rooms":10})
// @Success 200 {object} object "{ \"zone\": {} }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Router /zones/{id} [put]
func (h *ZoneHandler) UpdateZone(c *gin.Context) {
	zoneID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		MaxRooms    int    `json:"max_rooms" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	z, err := h.zoneManager.UpdateZone(zoneID, req.Name, req.Description, req.MaxRooms)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zone": z})
}

// DeleteZone 删除游戏区
// @Summary 删除游戏区
// @Description 删除指定游戏区（要求游戏区内没有房间）
// @Tags 游戏区
// @Produce json
// @Security BearerAuth
// @Param id path string true "游戏区 ID"
// @Success 200 {object} object "{ \"status\": \"ok\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Router /zones/{id} [delete]
func (h *ZoneHandler) DeleteZone(c *gin.Context) {
	zoneID := c.Param("id")

	if err := h.zoneManager.DeleteZone(zoneID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
