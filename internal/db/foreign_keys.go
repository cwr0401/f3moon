package db

import (
	"fmt"

	"gorm.io/gorm"
)

// foreignKeySpec 描述一个外键约束
type foreignKeySpec struct {
	name      string // 约束名
	table     string // 子表
	column    string // 子表列
	refTable  string // 父表
	refColumn string // 父表列
	onDelete  string // ON DELETE 行为: CASCADE | RESTRICT | SET NULL
	onUpdate  string // ON UPDATE 行为
}

// ApplyForeignKeys 在 AutoMigrate 之后应用外键约束。
// 通过原生 SQL 添加，避免在 GORM 模型上引入跨包关联字段，
// 从而避免 internal/room 反向依赖 internal/auth、internal/zone 等。
func ApplyForeignKeys(gdb *gorm.DB) error {
	specs := []foreignKeySpec{
		// verification_tokens.user_id -> users.id (用户删除时令牌一并删除)
		{
			name: "fk_verification_tokens_user", table: "verification_tokens", column: "user_id",
			refTable: "users", refColumn: "id",
			onDelete: "CASCADE", onUpdate: "CASCADE",
		},
		// rooms.zone_id -> game_zones.id (限制删除非空区)
		{
			name: "fk_rooms_zone", table: "rooms", column: "zone_id",
			refTable: "game_zones", refColumn: "id",
			onDelete: "RESTRICT", onUpdate: "CASCADE",
		},
		// rooms.owner -> users.id (限制删除拥有房间的用户)
		{
			name: "fk_rooms_owner", table: "rooms", column: "owner",
			refTable: "users", refColumn: "id",
			onDelete: "RESTRICT", onUpdate: "CASCADE",
		},
		// user_rooms.user_id -> users.id
		{
			name: "fk_user_rooms_user", table: "user_rooms", column: "user_id",
			refTable: "users", refColumn: "id",
			onDelete: "CASCADE", onUpdate: "CASCADE",
		},
		// user_rooms.room_id -> rooms.id
		{
			name: "fk_user_rooms_room", table: "user_rooms", column: "room_id",
			refTable: "rooms", refColumn: "id",
			onDelete: "CASCADE", onUpdate: "CASCADE",
		},
		// room_scores.room_id -> rooms.id
		{
			name: "fk_room_scores_room", table: "room_scores", column: "room_id",
			refTable: "rooms", refColumn: "id",
			onDelete: "CASCADE", onUpdate: "CASCADE",
		},
		// room_scores.player_id -> users.id
		{
			name: "fk_room_scores_player", table: "room_scores", column: "player_id",
			refTable: "users", refColumn: "id",
			onDelete: "CASCADE", onUpdate: "CASCADE",
		},
		// game_decks.room_id -> rooms.id
		{
			name: "fk_game_decks_room", table: "game_decks", column: "room_id",
			refTable: "rooms", refColumn: "id",
			onDelete: "CASCADE", onUpdate: "CASCADE",
		},
	}

	for _, s := range specs {
		if err := ensureForeignKey(gdb, s); err != nil {
			return fmt.Errorf("apply foreign key %s: %w", s.name, err)
		}
	}
	return nil
}

// ensureForeignKey 检查外键是否存在，不存在则添加。
func ensureForeignKey(gdb *gorm.DB, s foreignKeySpec) error {
	var count int64
	if err := gdb.Raw(`
		SELECT COUNT(*)
		FROM information_schema.TABLE_CONSTRAINTS
		WHERE CONSTRAINT_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND CONSTRAINT_NAME = ?
		  AND CONSTRAINT_TYPE = 'FOREIGN KEY'`,
		s.table, s.name,
	).Scan(&count).Error; err != nil {
		return fmt.Errorf("check fk existence: %w", err)
	}
	if count > 0 {
		return nil
	}

	stmt := fmt.Sprintf(
		"ALTER TABLE `%s` ADD CONSTRAINT `%s` FOREIGN KEY (`%s`) REFERENCES `%s`(`%s`) ON DELETE %s ON UPDATE %s",
		s.table, s.name, s.column, s.refTable, s.refColumn, s.onDelete, s.onUpdate,
	)
	if err := gdb.Exec(stmt).Error; err != nil {
		return fmt.Errorf("exec alter table: %w", err)
	}
	return nil
}
