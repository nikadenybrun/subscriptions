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

type Storage struct {
	Db *pg.DB
}

func NewStorage(storagePath string, cfg config.DBSaver) (*Storage, error) {
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
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	obj := &Storage{
		Db: conn,
	}
	if err := migrateSchema(obj); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}
	return obj, nil
}

func migrateSchema(s *Storage) error {
	schemas := []interface{}{
		(*models.Post)(nil),
		(*models.Comment)(nil),
	}

	op := orm.CreateTableOptions{IfNotExists: true}

	for _, schema := range schemas {
		if err := s.Db.Model(schema).CreateTable(&op); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (s *Storage) CloseDB() error {
	return s.Db.Close()
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
