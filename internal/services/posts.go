package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	models "subscriptions/internal/domain"
	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/storage"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

type PostService interface {
	CreatePost(ctx context.Context, p *model.Post) (string, error)
	GetPost(ctx context.Context, id string) (*models.Post, error)
	GetAllPosts(ctx context.Context) ([]*models.Post, error)
}

type Posts struct {
	db *pg.DB
}

func NewPost(storage *pg.DB) *Posts {
	return &Posts{db: storage}
}

func (s *Posts) SavePost(ctx context.Context, post *model.Post) (string, error) {
	const op = "storage.postgres.SavePost"

	var err error
	entity := &models.Post{
		ID:              uuid.New().String(),
		Title:           post.Title,
		Content:         post.Content,
		CommentsAllowed: post.CommentsAllowed,
		CreatedAt:       post.CreatedAt,
	}
	for i := 0; i < storage.MaxRetries; i++ {
		err = s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			_, err := tx.Model(entity).Insert()
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

func (s *Posts) GetAll(ctx context.Context) (*[]models.Post, error) {
	var posts []models.Post
	op := "GetAll"
	var err error
	for i := 0; i < storage.MaxRetries; i++ {
		err := s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err = tx.Model(&posts).Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrNoPostsFound)
				}
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		})
		if err == nil {
			return &posts, nil
		}
		if !isRetryableError(err) {
			return nil, err
		}
		time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
	}
	return nil, fmt.Errorf("%s: exceeded max retries, last error: %w", op, err)
}

func (s *Posts) GetPost(ctx context.Context, id string) (*models.Post, error) {
	op := "GetPost"
	var err error
	var post models.Post
	for i := 0; i < storage.MaxRetries; i++ {
		err := s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			err = tx.Model(&post).Where("id = ?", id).Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrNoPostsFound)
				}
				return fmt.Errorf("%s: %w", op, err)
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

	return nil, fmt.Errorf("getting post failed after %d retries: %w", storage.MaxRetries, err)
}
