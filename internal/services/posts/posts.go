package posts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/storage"
	database "subscriptions/internal/storage/postgres"

	"github.com/go-pg/pg/v10"
)

const maxRetries = 5

type Posts struct {
	ID              string
	Title           string
	Content         string
	CommentsAllowed bool
	// Comments        []model.Comment
}

func (link *Posts) SavePost(ctx context.Context) (string, error) {
	const op = "storage.postgres.SavePost"

	var err error
	for i := 0; i < maxRetries; i++ {
		err = database.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			_, err := tx.Model(link).Insert()
			if err != nil {
				return fmt.Errorf("%s: failed to insert: %w", op, err)
			}
			return nil
		})
		if err == nil {
			return link.ID, nil
		}
		if !isRetryableError(err) {
			return "", err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return "", fmt.Errorf("insert failed after %d retries: %w", maxRetries, err)
}

func GetAll(ctx context.Context) ([]Posts, error) {
	var posts []Posts
	op := "GetAll"
	var err error
	for i := 0; i < maxRetries; i++ {
		err := database.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err := tx.Model(&posts).Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrNoPostsFound)
				}
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		})
		if err == nil {
			return posts, nil
		}
		if !isRetryableError(err) {
			return nil, err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return nil, fmt.Errorf("%s: exceeded max retries, last error: %w", op, err)
}

func GetPost(ctx context.Context, id string) (*Posts, error) {
	op := "GetPost"
	var err error
	var post Posts
	var comments []model.Comment
	for i := 0; i < maxRetries; i++ {
		err := database.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err := tx.Model(&post).Where("id = ?", id).Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrNoPostsFound)
				}
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		})
		if err == nil {
			break
		}
		if !isRetryableError(err) {
			return nil, err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	for i := 0; i < maxRetries; i++ {
		err := database.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err := tx.Model(&comments).Where("post_id = ?", post.ID).Order("created_at ASC").Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					comments = []model.Comment{}
				} else {
					return fmt.Errorf("failed to retrieve comments: %w", err)
				}
			}
			return nil
		})
		if err == nil {
			return &post, nil
		}
		if !isRetryableError(err) {
			return nil, err
		}

		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return nil, fmt.Errorf("update balance failed after %d retries: %w", maxRetries, err)
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
