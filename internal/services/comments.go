package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	models "subscriptions/internal/domain"
	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/storage"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

type Comments struct {
	db *pg.DB
}

func NewComment(storage *pg.DB) *Comments {
	return &Comments{db: storage}
}

type CommentService interface {
	CheckCommentId(ctx context.Context, comtId string, postId string) error
	GetComments(ctx context.Context, postId string, first *int32, after *string) (*[]models.Comment, string, bool, error)
	SaveComment(ctx context.Context, c *model.Comment) (string, error)
}

func (c *Comments) SaveComment(ctx context.Context, com *model.Comment) (string, error) {
	const op = "storage.postgres.SavePost"
	if c == nil {
		return "", errors.New("Comments service is nil")
	}
	if c.db == nil {
		return "", errors.New("Comments service storage is nil")
	}

	createdAt := time.Now()
	entity := &models.Comment{
		ID:     uuid.New().String(),
		PostID: com.PostID,
		Text:   com.Text,

		CreatedAt: createdAt,
	}
	if com.ParentID != nil {
		entity.ParentID = *com.ParentID
	}
	var err error
	for i := 0; i < storage.MaxRetries; i++ {
		err = c.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			_, err = tx.Model(entity).Insert()
			if err != nil {
				return fmt.Errorf("%s: failed to insert: %w", op, err)
			}
			return nil
		})

		if err == nil {
			return entity.ID, nil
		}

		if !isRetryableError(err) {
			return "", err
		}

		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}

	return "", fmt.Errorf("insert failed after %d retries: %w", storage.MaxRetries, err)
}

func (c *Comments) GetComments(ctx context.Context, postId string, first *int32, after *string) (*[]models.Comment, string, bool, error) {
	op := "GetComments"
	var err error
	var coms []models.Comment
	var endCursor string
	for i := 0; i < storage.MaxRetries; i++ {
		err = c.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			if first == nil {
				return fmt.Errorf("You need to add param first", op)
			}
			if *first != 0 {
				query := tx.Model(&coms).Where("post_id = ?", postId).Order("created_at ASC").Limit(int(*first))
				if after != nil && *after != "" {
					var afterComment models.Comment
					err = tx.Model(&afterComment).Where("id = ?", *after).Select()
					if err != nil {
						if errors.Is(err, pg.ErrNoRows) {
							return fmt.Errorf("invalid cursor: %w", err, op)
						}
						return err
					}
					query = query.Where("(created_at, id) > (?, ?)", afterComment.CreatedAt, afterComment.ID)
				}

				err = query.Select()
				if err != nil {
					return err
				}
				if len(coms) > 0 {
					endCursor = coms[len(coms)-1].ID
				}
				return nil
			}
			return nil
		})
		if err == nil {
			hasNextPage := len(coms) == int(*first)
			return &coms, endCursor, hasNextPage, nil
		}
		if !isRetryableError(err) {
			return nil, "", false, err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return nil, "", false, fmt.Errorf("GetComments failed after %d retries: %w", storage.MaxRetries, err, op)
}

func (c *Comments) CheckCommentId(ctx context.Context, comtId *string, postId string) error {
	op := "CheckCommentId"
	var err error
	var coms models.Comment
	if comtId == nil {
		return fmt.Errorf("%s: %w", op, storage.ErrCommentNotFound)
	}
	for i := 0; i < storage.MaxRetries; i++ {
		err = c.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err = tx.Model(&coms).Where("id = ?", *comtId).Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrCommentNotFound)
				}
				return fmt.Errorf("%s: %w", op, err)
			}

			if coms.PostID != postId {
				return fmt.Errorf("%s: this comment is not under this post", op)
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
	return fmt.Errorf("CheckCommentId failed after %d retries: %w", storage.MaxRetries, err)
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
