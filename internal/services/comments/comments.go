package comments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/storage"
	"subscriptions/internal/storage/postgres"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

const maxRetries = 5

type Comments struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postID"`
	Text      string    `json:"text"`
	ParentID  string    `json:"parentID"`
	CreatedAt time.Time `json:"createdAt"`
}

func (com *Comments) SaveComment(ctx context.Context, s *postgres.Storage, c *model.Comment) (string, error) {
	const op = "storage.postgres.SavePost"
	com.PostID, com.Text, com.ParentID, com.CreatedAt = c.PostID, c.Text, *c.ParentID, c.CreatedAt
	com.ID = uuid.New().String()
	var err error
	for i := 0; i < maxRetries; i++ {
		err = s.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			_, err := tx.Model(com).Insert()
			if err != nil {
				return fmt.Errorf("%s: failed to insert: %w", op, err)
			}
			return nil
		})

		if err == nil {
			return com.ID, nil
		}

		if !isRetryableError(err) {
			return "", err
		}

		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}

	return "", fmt.Errorf("insert failed after %d retries: %w", maxRetries, err)
}

func (com *Comments) GetComments(ctx context.Context, s *postgres.Storage, postId string, first *int32, after *string) ([]Comments, string, bool, error) {
	op := "GetComments"
	var err error
	var coms []Comments
	var endCursor string
	for i := 0; i < maxRetries; i++ {
		err := s.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			query := tx.Model(&coms).Where("post_id = ?", postId).Order("created_at ASC").Limit(int(*first))
			if after != nil && *after != "" {
				var afterComment Comments
				err := tx.Model(&afterComment).Where("id = ?", *after).Select()
				if err != nil {
					if errors.Is(err, pg.ErrNoRows) {
						return fmt.Errorf("invalid cursor: %w", err)
					}
					return err
				}
				query = query.Where("created_at > ?", afterComment.CreatedAt)
			}

			err := query.Select()
			if err != nil {
				return err
			}
			if len(coms) > 0 {
				endCursor = coms[len(coms)-1].ID
			}
			return nil
		})
		if err == nil {
			hasNextPage := len(coms) == int(*first)
			return coms, endCursor, hasNextPage, nil
		}
		if !isRetryableError(err) {
			return nil, "", false, err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return nil, "", false, fmt.Errorf("update balance failed after %d retries: %w", maxRetries, err, op)
}

func (com *Comments) CheckCommentId(ctx context.Context, s *postgres.Storage, comtId string) error {
	op := "CheckCommentId"
	var err error
	var coms Comments
	for i := 0; i < maxRetries; i++ {
		err := s.Db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err := tx.Model(&coms).Where("id = ?", comtId).Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrCommentNotFound)
				}
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		})
		if err == nil {
			return nil
		}
		if !isRetryableError(err) {
			return err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return fmt.Errorf("update balance failed after %d retries: %w", maxRetries, err, op)
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
