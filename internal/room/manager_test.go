package room

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cwr0401/f3moon/internal/model"
)

type recordingRepository struct {
	created []*Room
	updated []*Room
	closed  []string
	scores  map[string]int
}

func newRecordingRepository() *recordingRepository {
	return &recordingRepository{scores: make(map[string]int)}
}

func (r *recordingRepository) CreateRoom(room *Room) error {
	r.created = append(r.created, room)
	return nil
}

func (r *recordingRepository) UpdateRoom(room *Room) error {
	r.updated = append(r.updated, room)
	return nil
}

func (r *recordingRepository) CloseRoom(roomID string, closedAt time.Time) error {
	r.closed = append(r.closed, roomID)
	return nil
}

func (r *recordingRepository) UpsertScore(roomID, playerID string, score int) error {
	r.scores[roomID+":"+playerID] = score
	return nil
}

func TestManagerCreateRoomUsesUUIDAndRecordsMetadata(t *testing.T) {
	repo := newRecordingRepository()
	manager := NewManager(repo)

	r, err := manager.CreateRoom("花牌局", model.GameMode4Player, "owner-1")
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
	manager := NewManager(newRecordingRepository())

	r, err := manager.CreateRoom("三人局", model.GameMode3Player, "owner-1")
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}

	if r.MaxPlayers != 3 {
		t.Fatalf("max players = %d, want 3", r.MaxPlayers)
	}
}

func TestRoomMutationsUpdateTimestampsAndScores(t *testing.T) {
	repo := newRecordingRepository()
	manager := NewManager(repo)
	r, err := manager.CreateRoom("花牌局", model.GameMode4Player, "owner-1")
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
	manager := NewManager(repo)
	r, err := manager.CreateRoom("花牌局", model.GameMode4Player, "owner-1")
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}

	if err := manager.CloseRoom(r.ID); err != nil {
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
