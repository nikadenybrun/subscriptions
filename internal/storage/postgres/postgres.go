package postgres

import (
	"context"
	"fmt"
	"strings"
	"subscriptions/internal/config"
	models "subscriptions/internal/domain"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

const maxRetries = 5

var Db *pg.DB

func InitDB(storagePath string, cfg config.DBSaver) error {
	const op = "storage.postgres.New"

	conn := pg.Connect(&pg.Options{
		Addr:     cfg.DbAddr,
		User:     cfg.DbUser,
		Password: cfg.DbPass,
		Database: cfg.DbName,
	})
	var err error
	for i := 0; i < 10; i++ {
		if err = conn.Ping(context.Background()); err != nil {
			break

		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	Db = conn
	if err := migrateSchema(); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}
	return nil
}

func migrateSchema() error {
	schemas := []interface{}{
		(*models.Posts)(nil),
		(*models.Comments)(nil),
	}

	op := orm.CreateTableOptions{IfNotExists: true}

	for _, schema := range schemas {
		if err := Db.Model(schema).CreateTable(&op); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func CloseDB() error {
	return Db.Close()
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, "deadlock detected") ||
		strings.Contains(errMsg, "could not serialize access") ||
		strings.Contains(errMsg, "canceling statement due to conflict") ||
		strings.Contains(errMsg, "timeout") {
		return true
	}
	return false
}
