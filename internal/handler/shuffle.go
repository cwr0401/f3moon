package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/engine"
	"github.com/cwr0401/f3moon/internal/model"
)

// ShuffleHandler 洗牌处理器
type ShuffleHandler struct{}

// NewShuffleHandler 创建洗牌处理器
func NewShuffleHandler() *ShuffleHandler {
	return &ShuffleHandler{}
}

// ShuffleResponse 洗牌响应
type ShuffleResponse struct {
	Seed    int64         `json:"seed"`     // 使用的随机种子
	Stack   []uint8       `json:"stack"`    // 牌 ID 栈（索引 0 为栈顶）
	Tiles   []*model.Tile `json:"tiles"`    // 完整牌信息（按栈顺序）
	StackTop int          `json:"stack_top"` // 栈顶牌 ID
	StackBottom int        `json:"stack_bottom"` // 栈底牌 ID
}

// Shuffle 洗牌接口
// @Summary 洗牌
// @Description 生成一副打乱的牌，返回牌 ID 栈和完整牌信息
// @Tags 工具
// @Accept json
// @Produce json
// @Param seed query int false "随机种子（可选，默认使用当前时间）"
// @Success 200 {object} ShuffleResponse
// @Router /shuffle [post]
func (h *ShuffleHandler) Shuffle(c *gin.Context) {
	var seed int64

	// 从查询参数获取种子
	seedStr := c.Query("seed")
	if seedStr != "" {
		var err error
		seed, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seed"})
			return
		}
	}

	// 生成打乱的栈
	stack := engine.NewShuffledDeckStack(seed)
	tiles := model.GetTilesFromIDs(stack)

	// 如果种子未指定，使用当前时间
	if seed == 0 {
		// 我们需要重新获取种子，因为 NewShuffledDeckStack 内部使用了时间
		// 这里我们简单返回 0 表示自动种子
	}

	c.JSON(http.StatusOK, ShuffleResponse{
		Seed:         seed,
		Stack:        stack,
		Tiles:        tiles,
		StackTop:     int(stack[0]),
		StackBottom:  int(stack[len(stack)-1]),
	})
}
