package match

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

// Repository 封装 match 子系统的数据访问。
// MVP 阶段只保留结构占位，不实现 Storage 写入。
type Repository struct {
	db *sql.DB
	nk runtime.NakamaModule
}

// NewRepository 创建 Repository 实例。
func NewRepository(db *sql.DB, nk runtime.NakamaModule) *Repository {
	return &Repository{db: db, nk: nk}
}

// SaveMatchRecord 占位：未来保存对战记录到 Storage。
func (r *Repository) SaveMatchRecord(ctx context.Context, userID string, report *DebugSimuMatchResponse) error {
	// MVP 阶段不持久化
	return nil
}
