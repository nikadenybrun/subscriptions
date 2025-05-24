package graph

import (
	"context"
	models "subscriptions/internal/domain"
	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/lib/locks"

	"golang.org/x/exp/slog"
)

type Resolver struct {
	Logger                   *slog.Logger
	CommentAddedNotification chan *model.Comment
	Storage                  Storage
	Post_                    PostInterface
	Comment_                 CommentInterface
	Lock                     *locks.Locks
}
type Storage interface {
	CloseDB() error
}

//go:generate go run github.com/vektra/mockery/v2@v2.40.2 --name=PostInterface
type PostInterface interface {
	SavePost(ctx context.Context, p *model.Post) (string, error)
	GetAll(ctx context.Context) (*[]models.Post, error)
	GetPost(ctx context.Context, id string) (*models.Post, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.40.2 --name=CommentInterface
type CommentInterface interface {
	SaveComment(ctx context.Context, c *model.Comment) (string, error)
	GetComments(ctx context.Context, postId string, first *int32, after *string) (*[]models.Comment, string, bool, error)
	CheckCommentId(ctx context.Context, comtId *string, postId string) error
}
