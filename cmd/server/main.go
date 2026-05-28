package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/config"
	"github.com/cwr0401/f3moon/internal/handler"
	"github.com/cwr0401/f3moon/internal/room"
	"github.com/cwr0401/f3moon/internal/ws"
)

func main() {
	cfg := config.DefaultConfig()

	// 初始化组件
	roomManager := room.NewManager()
	hub := ws.NewHub()
	roomHandler := handler.NewRoomHandler(roomManager)
	gameHandler := handler.NewGameHandler()
	wsHandler := handler.NewWSHandler(hub)

	// 设置Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 中间件
	r.Use(corsMiddleware())

	// 静态文件(前端)
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("./web/templates/*")

	// 页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API路由
	api := r.Group("/api/v1")
	{
		// 房间
		rooms := api.Group("/rooms")
		{
			rooms.POST("", roomHandler.CreateRoom)
			rooms.GET("", roomHandler.ListRooms)
			rooms.GET("/:id", roomHandler.GetRoom)
			rooms.POST("/:id/join", roomHandler.JoinRoom)
			rooms.POST("/:id/leave", roomHandler.LeaveRoom)
			rooms.POST("/:id/ai", roomHandler.AddAIPlayer)
			rooms.POST("/:id/start", roomHandler.StartGame)
		}

		// 游戏
		games := api.Group("/games")
		{
			games.GET("/:id", gameHandler.GetGame)
			games.POST("/:id/cut", gameHandler.Cut)
			games.POST("/:id/tong", gameHandler.Tong)
			games.POST("/:id/draw", gameHandler.Draw)
			games.POST("/:id/discard", gameHandler.Discard)
			games.POST("/:id/pair", gameHandler.Pair)
			games.POST("/:id/ganta", gameHandler.Ganta)
			games.POST("/:id/win", gameHandler.Win)
			games.POST("/:id/pass", gameHandler.Pass)
			games.POST("/:id/dang-jing", gameHandler.DangJing)
		}
	}

	// WebSocket
	r.GET("/ws", wsHandler.HandleWebSocket)

	// 启动服务
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	log.Printf("荆楚花牌服务端启动于 %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
