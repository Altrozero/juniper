package juniper

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type ModelId int

// String implements fmt.Stringer
func (i ModelId) String() string {
	return strconv.Itoa(int(i))
}

var _ fmt.Stringer = (*ModelId)(nil)

// Model is the basis for a database model
type Model struct {
	ID        ModelId `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SoftDeleteModel is the basis for a database model with soft delete functionality
type SoftDeleteModel struct {
	ID        ModelId `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// PreloadDBModels updates the db transaction with assigned model preloads
func PreloadDBModels(db *gorm.DB, preload []string) *gorm.DB {
	if len(preload) == 0 {
		return db
	}

	tx := db
	for _, assoc := range preload {
		tx = tx.Preload(assoc)
	}

	return tx
}
