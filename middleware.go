package juniper

import (
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

const gormTransactionKey = "gorm_tx"

// TxHandle gets the gorm transaction handler from the gin context
func TxHandle(ctx context.Context) *gorm.DB {
	db, _ := ctx.Value(gormTransactionKey).(*gorm.DB)
	return db
}

// DBTransactionMiddleware handles the create, commit and rollback steps of a gorm transaction
// for each gin http request
func DBTransactionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tx := db.Begin()

		defer func() {
			if err := recover(); err != nil {
				tx.Rollback()
			}
		}()

		ctx.Set(gormTransactionKey, tx)
		ctx.Next()

		if ctx.Writer.Status() < 200 || ctx.Writer.Status() > 299 {
			tx.Rollback()
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("failed to commit transaction: %s", err)
		}
	}
}
