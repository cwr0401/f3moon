package room

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cwr0401/f3moon/internal/model"
	"github.com/cwr0401/f3moon/internal/zone"
)

type recordingRepository struct {
	created     []*Room
	updated     []*Room
	closed      []string
	scores      map[string]int
	userRooms   map[string]string
	activeRooms map[string]bool // roomID -> active
}

func newRecordingRepository() *recordingRepository {
	return &recordingRepository{
		scores:      make(map[string]int),
		userRooms:   make(map[string]string),
		activeRooms: make(map[string]bool),
	}
}

func (r *recordingRepository) CreateRoom(room *Room) error {
	r.created = append(r.created, room)
	r.activeRooms[room.ID] = true
	return nil
}

func (r *recordingRepository) UpdateRoom(room *Room) error {
	r.updated = append(r.updated, room)
	return nil
}

func (r *recordingRepository) CloseRoom(roomID string, closedAt time.Time) error {
	r.closed = append(r.closed, roomID)
	r.activeRooms[roomID] = false
	return nil
}

func (r *recordingRepository) UpsertScore(roomID, playerID string, score int) error {
	r.scores[roomID+":"+playerID] = score
	return nil
}

func (r *recordingRepository) GetUserRoom(userID string) (string, error) {
	return r.userRooms[userID], nil
}

func (r *recordingRepository) SetUserRoom(userID, roomID string) error {
	r.userRooms[userID] = roomID
	return nil
}

func (r *recordingRepository) RemoveUserRoom(userID string) error {
	delete(r.userRooms, userID)
	return nil
}

func (r *recordingRepository) IsRoomActive(roomID string) (bool, error) {
	return r.activeRooms[roomID], nil
}

func (r *recordingRepository) CountActiveRoomsByZone(zoneID string) (int, error) {
	count := 0
	for _, room := range r.created {
		if room.ZoneID == zoneID && r.activeRooms[room.ID] {
			count++
		}
	}
	return count, nil
}

// mockZoneRepository is a mock implementation of zone.Repository
type mockZoneRepository struct {
	zones map[string]*zone.GameZone
}

func newMockZoneRepository() *mockZoneRepository {
	return &mockZoneRepository{
		zones: make(map[string]*zone.GameZone),
	}
}

func (m *mockZoneRepository) CreateZone(z *zone.GameZone) error {
	m.zones[z.ID] = z
	return nil
}

func (m *mockZoneRepository) UpdateZone(z *zone.GameZone) error {
	m.zones[z.ID] = z
	return nil
}

func (m *mockZoneRepository) DeleteZone(zoneID string) error {
	delete(m.zones, zoneID)
	return nil
}

func (m *mockZoneRepository) GetZone(zoneID string) (*zone.GameZone, error) {
	return m.zones[zoneID], nil
}

func (m *mockZoneRepository) GetZoneByName(name string) (*zone.GameZone, error) {
	for _, z := range m.zones {
		if z.Name == name {
			return z, nil
		}
	}
	return nil, nil
}

func (m *mockZoneRepository) ListZones() ([]*zone.GameZone, error) {
	zones := make([]*zone.GameZone, 0, len(m.zones))
	for _, z := range m.zones {
		zones = append(zones, z)
	}
	return zones, nil
}

func setupTestZoneManager() (*zone.Manager, string) {
	repo := newMockZoneRepository()
	zm := zone.NewManager(repo)
	// Create a test zone
	z, err := zm.CreateZone("测试区", "用于测试", 100)
	if err != nil {
		panic(err)
	}
	return zm, z.ID
}

func TestManagerCreateRoomUsesUUIDAndRecordsMetadata(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)

	r, err := manager.CreateRoom(zoneID, "花牌局", model.GameMode4Player, "owner-1", 8)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}

	if _, err := uuid.Parse(r.ID); err != nil {
		t.Fatalf("room ID should be UUID, got %q: %v", r.ID, err)
	}
	if r.CreatedAt.IsZero() {
		t.Fatal("created_at should be set")
	}
	if r.UpdatedAt.IsZero() {
		t.Fatal("updated_at should be set")
	}
	if !r.CreatedAt.Equal(r.UpdatedAt) {
		t.Fatalf("new room updated_at should equal created_at, got created=%s updated=%s", r.CreatedAt, r.UpdatedAt)
	}
	if r.ClosedAt != nil {
		t.Fatalf("new room closed_at should be nil, got %v", r.ClosedAt)
	}
	if r.Mode != model.GameMode4Player {
		t.Fatalf("mode = %v, want %v", r.Mode, model.GameMode4Player)
	}
	if r.MaxPlayers != 4 {
		t.Fatalf("max players = %d, want 4", r.MaxPlayers)
	}
	if r.Scores == nil {
		t.Fatal("scores should be initialized")
	}
	if len(repo.created) != 1 || repo.created[0] != r {
		t.Fatalf("repository should record created room once, got %d", len(repo.created))
	}
}

func TestManagerCreateRoomUsesThreePlayerCapacity(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)

	r, err := manager.CreateRoom(zoneID, "三人局", model.GameMode3Player, "owner-1", 8)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}

	if r.MaxPlayers != 3 {
		t.Fatalf("max players = %d, want 3", r.MaxPlayers)
	}
}

func TestRoomMutationsUpdateTimestampsAndScores(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)
	r, err := manager.CreateRoom(zoneID, "花牌局", model.GameMode4Player, "owner-1", 8)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	createdAt := r.UpdatedAt

	time.Sleep(time.Nanosecond)

	if ok := r.AddPlayer(&RoomPlayer{ID: "p1", Name: "玩家1"}); !ok {
		t.Fatal("expected AddPlayer to succeed")
	}
	if !r.UpdatedAt.After(createdAt) {
		t.Fatalf("AddPlayer should update UpdatedAt, before=%s after=%s", createdAt, r.UpdatedAt)
	}

	updatedAt := r.UpdatedAt
	time.Sleep(time.Nanosecond)
	r.SetReady("p1", true)
	if !r.UpdatedAt.After(updatedAt) {
		t.Fatalf("SetReady should update UpdatedAt, before=%s after=%s", updatedAt, r.UpdatedAt)
	}

	if err := manager.UpdateScore(r.ID, "p1", 12); err != nil {
		t.Fatalf("UpdateScore returned error: %v", err)
	}
	if r.Scores["p1"] != 12 {
		t.Fatalf("score = %d, want 12", r.Scores["p1"])
	}
	if repo.scores[r.ID+":p1"] != 12 {
		t.Fatalf("repository score = %d, want 12", repo.scores[r.ID+":p1"])
	}
}

func TestManagerCloseRoomRecordsClosedAt(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)
	r, err := manager.CreateRoom(zoneID, "花牌局", model.GameMode4Player, "owner-1", 8)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}

	if err := manager.CloseRoom(r.ID, "owner-1"); err != nil {
		t.Fatalf("CloseRoom returned error: %v", err)
	}

	if r.Status != RoomFinished {
		t.Fatalf("status = %v, want %v", r.Status, RoomFinished)
	}
	if r.ClosedAt == nil || r.ClosedAt.IsZero() {
		t.Fatalf("closed_at should be set, got %v", r.ClosedAt)
	}
	if !r.UpdatedAt.Equal(*r.ClosedAt) {
		t.Fatalf("updated_at should equal closed_at, got updated=%s closed=%s", r.UpdatedAt, *r.ClosedAt)
	}
	if len(repo.closed) != 1 || repo.closed[0] != r.ID {
		t.Fatalf("repository should record closed room once, got %#v", repo.closed)
	}
}

func TestManagerCreateRoomAndJoin(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)

	owner := &RoomPlayer{ID: "owner-1", Name: "房主"}
	r, err := manager.CreateRoomAndJoin(zoneID, "测试局", model.GameMode3Player, owner, 8)
	if err != nil {
		t.Fatalf("CreateRoomAndJoin returned error: %v", err)
	}

	if _, err := uuid.Parse(r.ID); err != nil {
		t.Fatalf("room ID should be UUID, got %q: %v", r.ID, err)
	}
	if r.PlayerCount() != 1 {
		t.Fatalf("should have 1 player, got %d", r.PlayerCount())
	}
	if r.Players[0].ID != owner.ID {
		t.Fatalf("player ID = %q, want %q", r.Players[0].ID, owner.ID)
	}
	if r.Owner != owner.ID {
		t.Fatalf("owner ID = %q, want %q", r.Owner, owner.ID)
	}
	if repo.userRooms[owner.ID] != r.ID {
		t.Fatalf("user room should be recorded")
	}
}

func TestManagerConcurrentRoomCreation(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)

	const goroutines = 10
	var wg sync.WaitGroup
	rooms := make([]*Room, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			owner := &RoomPlayer{ID: fmt.Sprintf("owner-%d", idx), Name: fmt.Sprintf("房主%d", idx)}
			r, err := manager.CreateRoomAndJoin(zoneID, fmt.Sprintf("测试局%d", idx), model.GameMode3Player, owner, 8)
			rooms[idx] = r
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: error = %v", i, err)
		}
		if rooms[i] == nil {
			t.Errorf("goroutine %d: room is nil", i)
		}
	}

	seenIDs := make(map[string]bool)
	for _, r := range rooms {
		if r != nil {
			if seenIDs[r.ID] {
				t.Errorf("duplicate room ID: %q", r.ID)
			}
			seenIDs[r.ID] = true
		}
	}
}

func TestManagerConcurrentJoins(t *testing.T) {
	repo := newRecordingRepository()
	zoneManager, zoneID := setupTestZoneManager()
	manager := NewManager(repo, zoneManager)

	owner := &RoomPlayer{ID: "owner-1", Name: "房主"}
	r, err := manager.CreateRoomAndJoin(zoneID, "测试局", model.GameMode4Player, owner, 8)
	if err != nil {
		t.Fatalf("CreateRoomAndJoin returned error: %v", err)
	}

	const goroutines = 3
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			player := &RoomPlayer{ID: fmt.Sprintf("player-%d", idx), Name: fmt.Sprintf("玩家%d", idx)}
			_, err := manager.JoinRoom(r.ID, player)
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: error = %v", i, err)
		}
	}

	if len(r.Players) != 4 {
		t.Errorf("should have 4 players, got %d", len(r.Players))
	}
}
