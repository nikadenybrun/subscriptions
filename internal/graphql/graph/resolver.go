package graph

import (
	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/services/comments"
	"subscriptions/internal/services/posts"

	"golang.org/x/exp/slog"
)

type Resolver struct {
	Logger                   *slog.Logger
	CommentAddedNotification chan *model.Comment
	Post_                    posts.Posts
	Comment_                 comments.Comments
}
