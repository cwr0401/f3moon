package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/middleware"
	"github.com/cwr0401/f3moon/internal/model"
)

// mockRepository 是 game.Repository 的 mock 实现
type mockRepository struct {
	deck *game.GameDeckRecord
}

func newMockRepository() *mockRepository {
	return &mockRepository{}
}

func (m *mockRepository) CreateGameDeck(deck *game.GameDeckRecord) error {
	m.deck = deck
	return nil
}

func (m *mockRepository) GetGameDeck(id string) (*game.GameDeckRecord, error) {
	if m.deck != nil && m.deck.ID == id {
		return m.deck, nil
	}
	return nil, nil
}

func (m *mockRepository) GetGameDeckByRoomID(roomID string) (*game.GameDeckRecord, error) {
	if m.deck != nil && m.deck.RoomID == roomID {
		return m.deck, nil
	}
	return nil, nil
}

func (m *mockRepository) UpdateGameDeck(deck *game.GameDeckRecord) error {
	m.deck = deck
	return nil
}

// mockNotifier 是 game.NotifyBroadcaster 的 mock 实现
type mockNotifier struct{}

func (m *mockNotifier) Broadcast(msg game.NotifyMessage) {}

// setupTestRouter 设置测试路由
func setupTestRouter(handler *GameHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 设置测试中间件，模拟用户认证
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "test-user-1")
		c.Next()
	})

	// 游戏路由
	gameGroup := router.Group("/games")
	{
		gameGroup.POST("/:id/cut", handler.Cut)
		gameGroup.POST("/:id/deal", handler.Deal)
		gameGroup.POST("/:id/tong", handler.Tong)
	}

	return router
}

// createTestGame 创建一个测试用的游戏状态机
func createTestGame(t *testing.T, roomID string, repo *mockRepository) (*game.StateMachine, *game.GameDeckRecord) {
	// 创建初始牌栈
	stack := make([]uint8, 112)
	for i := 0; i < 112; i++ {
		stack[i] = uint8(i)
	}

	// 创建游戏牌记录
	deck := &game.GameDeckRecord{
		ID:           uuid.NewString(),
		RoomID:       roomID,
		ShuffleStack: stack,
		CutStack:     stack,
		StackTop:     0,
		StackBottom:  111,
		Shuffled:     false,
		CutFinished:  false,
		DealFinished: false,
		TongFinished: false,
	}

	repo.deck = deck

	// 创建游戏状态
	gs := model.NewGameState(roomID, model.GameMode3Player)
	gs.Players[0] = &model.Player{ID: "player-0", Name: "庄家"}
	gs.Players[1] = &model.Player{ID: "player-1", Name: "闲家1"}
	gs.Players[2] = &model.Player{ID: "player-2", Name: "闲家2"}

	notifier := &mockNotifier{}
	sm := game.NewStateMachine(gs, notifier)

	return sm, deck
}

func TestCutDeck(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 先洗牌
	sm, deck := createTestGame(t, roomID, repo)
	deck.Shuffled = true
	gs := sm.Game()
	gs.Phase = model.PhaseCut // 设置正确的Phase
	gs.Players[2].ID = "test-user-1" // 设置当前用户为切牌人(闲家2)
	gameHandler.RegisterGame(roomID, sm)

	// 测试切牌接口
	cutReq := map[string]int{"position": 73}
	reqBody, _ := json.Marshal(cutReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/cut", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}

	if !repo.deck.CutFinished {
		t.Error("expected cut_finished to be true")
	}
	if repo.deck.CutPosition != 73 {
		t.Errorf("expected cut position 73, got %d", repo.deck.CutPosition)
	}
}

func TestCutDeck_InvalidPosition(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 先洗牌
	sm, _ := createTestGame(t, roomID, repo)
	repo.deck.Shuffled = true
	gameHandler.RegisterGame(roomID, sm)

	// 测试无效的切牌位置（太小）
	cutReq := map[string]int{"position": 30}
	reqBody, _ := json.Marshal(cutReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/cut", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDealCards(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 先洗牌和切牌
	sm, _ := createTestGame(t, roomID, repo)
	repo.deck.Shuffled = true
	repo.deck.CutFinished = true
	repo.deck.CutPosition = 73
	gs := sm.Game()
	gs.Players[0].ID = "test-user-1" // 设置当前用户为庄家
	gameHandler.RegisterGame(roomID, sm)

	// 测试发牌接口
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/deal", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}

	if !repo.deck.DealFinished {
		t.Error("expected deal_finished to be true")
	}
	dealerState := repo.deck.Players.FindByRole(0)
	if dealerState == nil || len(dealerState.Hand) != 26 {
		t.Errorf("expected dealer hand len 26, got %v", dealerState)
	}
	player1State := repo.deck.Players.FindByRole(1)
	if player1State == nil || len(player1State.Hand) != 25 {
		t.Errorf("expected player1 hand len 25, got %v", player1State)
	}
	player2State := repo.deck.Players.FindByRole(2)
	if player2State == nil || len(player2State.Hand) != 25 {
		t.Errorf("expected player2 hand len 25, got %v", player2State)
	}
}

func TestDealCards_BeforeShuffle(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 先不洗牌
	sm, _ := createTestGame(t, roomID, repo)
	gameHandler.RegisterGame(roomID, sm)

	// 尝试发牌
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/deal", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTong_Pass(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 设置完整的游戏状态
	sm, deck := createTestGame(t, roomID, repo)
	deck.Shuffled = true
	deck.CutFinished = true
	deck.DealFinished = true

	// 发牌 - 给玩家一些手牌
	dealerHand := make([]uint8, 26)
	player1Hand := make([]uint8, 25)
	player2Hand := make([]uint8, 25)
	for i := 0; i < 26; i++ {
		dealerHand[i] = uint8(i)
	}
	for i := 0; i < 25; i++ {
		player1Hand[i] = uint8(i + 26)
		player2Hand[i] = uint8(i + 51)
	}
	repo.deck.Players = game.PlayerDeckStates{
		{ID: "player-0", Role: 0, Hand: dealerHand, Finished: true},
		{ID: "player-1", Role: 1, Hand: player1Hand, Finished: true},
		{ID: "player-2", Role: 2, Hand: player2Hand, Finished: true},
	}
	repo.deck.StackTop = 76
	repo.deck.StackBottom = 111

	// 注册游戏并设置内存状态
	gameHandler.RegisterGame(roomID, sm)
	gs := sm.Game()
	gs.Phase = model.PhaseTongAsk // 设置正确的Phase
	gs.Players[0].Hand = model.GetTilesFromIDs(dealerHand)
	gs.Players[1].Hand = model.GetTilesFromIDs(player1Hand)
	gs.Players[2].Hand = model.GetTilesFromIDs(player2Hand)
	gs.Players[2].ID = "test-user-1" // 让当前用户是闲家2
	gs.DrawPile = model.GetTilesFromIDs(repo.deck.CutStack[repo.deck.StackTop : repo.deck.StackBottom+1])

	// 测试跳过统牌
	tongReq := map[string]interface{}{"skip": true}
	reqBody, _ := json.Marshal(tongReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/tong", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if repo.deck.TongCurrent != 1 {
		t.Errorf("expected tong_current 1, got %d", repo.deck.TongCurrent)
	}
}

func TestTong_InvalidTongSize(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 设置完整的游戏状态
	sm, _ := createTestGame(t, roomID, repo)
	repo.deck.Shuffled = true
	repo.deck.CutFinished = true
	repo.deck.DealFinished = true
	repo.deck.Players = game.PlayerDeckStates{
		{ID: "player-0", Role: 0, Hand: make([]uint8, 26), Finished: true},
		{ID: "player-1", Role: 1, Hand: make([]uint8, 25), Finished: true},
		{ID: "player-2", Role: 2, Hand: make([]uint8, 25), Finished: true},
	}

	gameHandler.RegisterGame(roomID, sm)
	gs := sm.Game()
	gs.Players[2].ID = "test-user-1"

	// 测试无效的统牌大小
	tongReq := map[string]interface{}{
		"tile_name": "三",
		"tong_size": 6,
		"skip":      false,
	}
	reqBody, _ := json.Marshal(tongReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/tong", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTong_BeforeDeal(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 只洗牌和切牌，不发牌
	sm, _ := createTestGame(t, roomID, repo)
	repo.deck.Shuffled = true
	repo.deck.CutFinished = true
	gameHandler.RegisterGame(roomID, sm)

	// 尝试统牌
	tongReq := map[string]interface{}{"skip": true}
	reqBody, _ := json.Marshal(tongReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/tong", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTong_NotYourTurn(t *testing.T) {
	repo := newMockRepository()
	gameHandler := NewGameHandler(repo)
	router := setupTestRouter(gameHandler)

	roomID := uuid.NewString()

	// 设置完整的游戏状态
	sm, _ := createTestGame(t, roomID, repo)
	repo.deck.Shuffled = true
	repo.deck.CutFinished = true
	repo.deck.DealFinished = true
	repo.deck.Players = game.PlayerDeckStates{
		{ID: "player-0", Role: 0, Hand: make([]uint8, 26), Finished: true},
		{ID: "player-1", Role: 1, Hand: make([]uint8, 25), Finished: true},
		{ID: "player-2", Role: 2, Hand: make([]uint8, 25), Finished: true},
	}

	gameHandler.RegisterGame(roomID, sm)
	gs := sm.Game()
	gs.Players[2].ID = "another-user" // 当前用户不是闲家2

	// 尝试统牌
	tongReq := map[string]interface{}{"skip": true}
	reqBody, _ := json.Marshal(tongReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/games/"+roomID+"/tong", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
