package graph

import (
	"context"
	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/services/comments"
	"subscriptions/internal/services/posts"
	"subscriptions/internal/storage/postgres"

	"golang.org/x/exp/slog"
)

type Resolver struct {
	Logger                   *slog.Logger
	CommentAddedNotification chan *model.Comment
	Storage                  *postgres.Storage
	Post_                    PostInterface
	Comment_                 CommentInterface
}

type PostInterface interface {
	GetAll(ctx context.Context, s *postgres.Storage) ([]posts.Posts, error)
	GetPost(ctx context.Context, s *postgres.Storage, id string) (*posts.Posts, error)
	SavePost(ctx context.Context, s *postgres.Storage, p *model.Post) (string, error)
}

type CommentInterface interface {
	SaveComment(ctx context.Context, s *postgres.Storage, c *model.Comment) (string, error)
	GetComments(ctx context.Context, s *postgres.Storage, postId string, first *int32, after *string) ([]comments.Comments, string, bool, error)
	CheckCommentId(ctx context.Context, s *postgres.Storage, comtId string) error
}
