package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/cwr0401/f3moon/docs"
	"github.com/cwr0401/f3moon/internal/auth"
	"github.com/cwr0401/f3moon/internal/config"
	"github.com/cwr0401/f3moon/internal/db"
	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/handler"
	"github.com/cwr0401/f3moon/internal/middleware"
	"github.com/cwr0401/f3moon/internal/room"
	"github.com/cwr0401/f3moon/internal/ws"
	"github.com/cwr0401/f3moon/internal/zone"
)

// @title 荆楚花牌 API
// @version 1.0
// @description 荆楚花牌(花好月圆)游戏服务端 API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.DefaultConfig()

	// 初始化数据库
	gdb, err := db.NewMySQL(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	if err := db.AutoMigrate(gdb, &auth.User{}, &auth.VerificationToken{}, &zone.GameZoneRecord{}, &room.RoomRecord{}, &room.RoomScore{}, &room.UserRoomRecord{}, &game.GameDeckRecord{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	if err := db.ApplyForeignKeys(gdb); err != nil {
		log.Fatalf("应用外键约束失败: %v", err)
	}
	log.Println("数据库连接成功")

	// 初始化认证组件
	jwtMgr, err := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)
	if err != nil {
		log.Fatalf("初始化JWT失败: %v", err)
	}

	repo := auth.NewGORMRepository(gdb)
	mailer := auth.NewMailer(auth.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		DevMode:  cfg.SMTP.DevMode,
	})
	authSvc := auth.NewService(repo, mailer, jwtMgr, cfg.AppBaseURL)

	// 初始化组件
	zoneRepo := zone.NewGORMRepository(gdb)
	zoneManager := zone.NewManager(zoneRepo)
	if err := zoneManager.LoadFromDB(); err != nil {
		log.Fatalf("加载游戏区失败: %v", err)
	}
	if err := zoneManager.InitializeDefaultZones(); err != nil {
		log.Fatalf("初始化默认游戏区失败: %v", err)
	}

	roomRepo := room.NewGORMRepository(gdb)
	roomManager := room.NewManager(roomRepo, zoneManager)
	if err := zoneManager.RebuildRoomCounts(roomRepo.CountActiveRoomsByZone); err != nil {
		log.Fatalf("重建游戏区房间计数失败: %v", err)
	}
	gameRepo := game.NewGORMRepository(gdb)
	hub := ws.NewHub()
	authHandler := handler.NewAuthHandler(authSvc)
	gameHandler := handler.NewGameHandler(gameRepo)
	shuffleHandler := handler.NewShuffleHandler()
	roomHandler := handler.NewRoomHandler(roomManager, gameHandler, hub, gameRepo)
	zoneHandler := handler.NewZoneHandler(zoneManager)
	wsHandler := handler.NewWSHandler(hub, jwtMgr)

	// 设置Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 中间件
	r.Use(corsMiddleware())

	// 前端: Go 模板(认证页面)
	r.LoadHTMLGlob("./web/templates/*")
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// 前端: Vue SPA 构建产物
	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/favicon.svg", "./web/dist/favicon.svg")

	// SPA fallback: 所有未匹配的页面路由返回 index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" || path == "/ws" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./web/dist/index.html")
	})

	// API路由
	api := r.Group("/api/v1")
	{
		// 公开工具接口
		api.POST("/shuffle", shuffleHandler.Shuffle)

		// 认证（公开）
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.GET("/verify", authHandler.Verify)
			authGroup.POST("/resend", authHandler.Resend)
		}

		// 受保护接口
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(jwtMgr))
		{
			// 当前用户
			protected.GET("/auth/me", authHandler.Me)

			// 游戏区
			zones := protected.Group("/zones")
			{
				zones.GET("", zoneHandler.ListZones)
				zones.GET("/:id", zoneHandler.GetZone)
				zones.POST("", zoneHandler.CreateZone)
				zones.PUT("/:id", zoneHandler.UpdateZone)
				zones.DELETE("/:id", zoneHandler.DeleteZone)
			}

			// 房间
			rooms := protected.Group("/rooms")
			{
				rooms.POST("", roomHandler.CreateRoom)
				rooms.GET("", roomHandler.ListRooms)
				rooms.GET("/:id", roomHandler.GetRoom)
				rooms.POST("/:id/join", roomHandler.JoinRoom)
				rooms.POST("/:id/leave", roomHandler.LeaveRoom)
				rooms.POST("/:id/close", roomHandler.CloseRoom)
				rooms.POST("/:id/ai", roomHandler.AddAIPlayer)
				rooms.POST("/:id/ready", roomHandler.Ready)
				rooms.POST("/:id/start", roomHandler.StartGame)
			}

			// 游戏
			games := protected.Group("/games")
			{
				games.GET("/:id", gameHandler.GetGame)
				games.POST("/:id/cut", gameHandler.Cut)
				games.POST("/:id/deal", gameHandler.Deal)
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
	}

	// WebSocket
	r.GET("/ws", wsHandler.HandleWebSocket)

	// Swagger API 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 输出路由信息
	log.Println("注册路由:")
	for _, route := range r.Routes() {
		log.Printf("  %-6s %s", route.Method, route.Path)
	}

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
