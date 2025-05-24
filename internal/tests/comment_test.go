package services_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	models "subscriptions/internal/domain"
	"subscriptions/internal/graphql/graph"
	generated "subscriptions/internal/graphql/graph"
	"subscriptions/internal/graphql/graph/mocks"
	"subscriptions/internal/graphql/graph/model"

	"subscriptions/internal/storage/postgres"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestMutationResolver_CreateComment(t *testing.T) {
	t.Run("should create comment correctly", func(t *testing.T) {
		mockComment := mocks.NewCommentInterface(t)
		mockPost := mocks.NewPostInterface(t)
		postID := "post-123"
		parentID := "parent-456"
		text := "This is a comment"
		mockPost.On("GetPost", mock.Anything, postID).Return(&models.Post{
			ID:              postID,
			Title:           "Title",
			Content:         "Content",
			CommentsAllowed: true,
		}, nil)

		mockComment.On("CheckCommentId", mock.Anything, &parentID, mock.Anything).Return(nil)
		mockComment.On("SaveComment", mock.Anything, mock.Anything).Return("comment-id-789", nil)
		resolver := &graph.Resolver{
			Post_:                    mockPost,
			Comment_:                 mockComment,
			Storage:                  new(postgres.Storage),
			Logger:                   slog.New(slog.NewTextHandler(io.Discard, nil)),
			CommentAddedNotification: make(chan *model.Comment, 1),
		}
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver})))

		var resp struct {
			CreateComment struct {
				ID       string
				PostID   string
				ParentID *string
				Text     string
			}
		}
		mutation := fmt.Sprintf(`
			mutation {
				createComment(postID: "%s", parentID: "%s", text: "%s") {
					id
					postID
					parentID
					text
				}
			}
		`, postID, parentID, text)

		c.MustPost(mutation, &resp)
		mockPost.AssertExpectations(t)
		mockComment.AssertExpectations(t)
		require.Equal(t, "comment-id-789", resp.CreateComment.ID)
		require.Equal(t, postID, resp.CreateComment.PostID)
		require.NotNil(t, resp.CreateComment.ParentID)
		require.Equal(t, parentID, *resp.CreateComment.ParentID)
		require.Equal(t, text, resp.CreateComment.Text)
	})
}

func TestComment_CheckCommentId(t *testing.T) {
	t.Run("should check comment id without error", func(t *testing.T) {
		mockComment := mocks.NewCommentInterface(t)
		ctx := context.Background()
		commentID := "comment-123"
		postId := "postId-123"
		mockComment.On("CheckCommentId", ctx, &commentID, postId).Return(nil)

		err := mockComment.CheckCommentId(ctx, &commentID, postId)

		require.NoError(t, err)
		mockComment.AssertExpectations(t)
	})
}

func TestComment_GetComments(t *testing.T) {
	t.Run("should get comments correctly", func(t *testing.T) {
		mockComment := mocks.NewCommentInterface(t)
		ctx := context.Background()
		postID := "post-123"
		var first *int32 = nil
		var after *string = nil
		mockComments := []models.Comment{
			{
				ID:        "c1",
				PostID:    postID,
				Text:      "Comment 1",
				ParentID:  "",
				CreatedAt: time.Now(),
			},
			{
				ID:        "c2",
				PostID:    postID,
				Text:      "Comment 2",
				ParentID:  "c1",
				CreatedAt: time.Now(),
			},
		}
		mockComment.On("GetComments", ctx, postID, first, after).
			Return(&mockComments, "cursor123", true, nil)
		comments, cursor, hasNext, err := mockComment.GetComments(ctx, postID, first, after)
		require.NoError(t, err)
		require.Equal(t, &mockComments, comments)
		require.Equal(t, "cursor123", cursor)
		require.True(t, hasNext)
		mockComment.AssertExpectations(t)
	})
}
