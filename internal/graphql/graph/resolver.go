package graph

import (
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
	Post_                    *posts.Posts
	Comment_                 *comments.Comments
}

// type PostInterface interface {
// 	SavePost(ctx context.Context, s *postgres.Storage) (string, error)
// 	GetAll(ctx context.Context, s *postgres.Storage) ([]posts.Posts, error)
// 	GetPost(ctx context.Context, s *postgres.Storage, id string) (*posts.Posts, error)
// }

// //go:generate go run github.com/vektra/mockery/v2@v2.40.2 --name=Comment
// type CommentInterface interface {
// 	SaveComment(ctx context.Context, s *postgres.Storage) (string, error)
// 	GetComments(ctx context.Context, s *postgres.Storage, postId string, first *int32, after *string) ([]comments.Comments, string, bool, error)
// 	CheckCommentId(ctx context.Context, s *postgres.Storage, comtId string) error
// }
