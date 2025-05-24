package services_test

import (
	"fmt"
	"io"
	models "subscriptions/internal/domain"
	"subscriptions/internal/graphql/graph"
	generated "subscriptions/internal/graphql/graph"
	"subscriptions/internal/graphql/graph/mocks"
	"subscriptions/internal/storage/postgres"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestMutationResolver_CreatePost(t *testing.T) {
	t.Run("should create post correctly", func(t *testing.T) {
		mockPost := mocks.NewPostInterface(t)
		mockPost.On("SavePost", mock.Anything, mock.Anything).Return("test-id", nil)
		resolver := &graph.Resolver{
			Post_:   mockPost,
			Storage: new(postgres.Storage),
		}
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver})))
		var resp struct {
			CreatePost struct {
				ID              string
				Title           string
				Content         string
				CommentsAllowed bool
			}
		}
		q := `
      mutation {
        createPost(title: "Test", content: "Content", commentsAllowed: true) {
          id
          title
          content
          commentsAllowed
        }
      }
    `
		c.MustPost(q, &resp)
		mockPost.AssertExpectations(t)
		require.Equal(t, "test-id", resp.CreatePost.ID)
		require.Equal(t, "Test", resp.CreatePost.Title)
		require.Equal(t, "Content", resp.CreatePost.Content)
		require.True(t, resp.CreatePost.CommentsAllowed)
	})
}
func TestQueryResolver_Posts(t *testing.T) {
	t.Run("should return all posts correctly", func(t *testing.T) {
		mockPost := mocks.NewPostInterface(t)

		mockPosts := []models.Post{
			{ID: "id1", Title: "Title1", Content: "Content1", CommentsAllowed: true},
			{ID: "id2", Title: "Title2", Content: "Content2", CommentsAllowed: false},
		}
		mockPost.On("GetAll", mock.Anything).Return(&mockPosts, nil)

		resolver := &graph.Resolver{
			Post_:   mockPost,
			Storage: new(postgres.Storage),
		}
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver})))

		var resp struct {
			Posts []struct {
				ID              string
				Title           string
				Content         string
				CommentsAllowed bool
			}
		}
		query := `
			query {
				posts {
					id
					title
					content
					commentsAllowed
				}
			}
		`
		c.MustPost(query, &resp)
		mockPost.AssertExpectations(t)
		require.Len(t, resp.Posts, 2)
		require.Equal(t, "id1", resp.Posts[0].ID)
		require.Equal(t, "Title1", resp.Posts[0].Title)
		require.Equal(t, "Content1", resp.Posts[0].Content)
		require.True(t, resp.Posts[0].CommentsAllowed)
		require.Equal(t, "id2", resp.Posts[1].ID)
		require.Equal(t, "Title2", resp.Posts[1].Title)
		require.Equal(t, "Content2", resp.Posts[1].Content)
		require.False(t, resp.Posts[1].CommentsAllowed)
	})
}

func TestQueryResolver_Post(t *testing.T) {
	t.Run("should return post with comments correctly", func(t *testing.T) {
		mockPost := mocks.NewPostInterface(t)
		mockComment := mocks.NewCommentInterface(t)

		postID := "test-post-id"
		mockPost.On("GetPost", mock.Anything, postID).Return(&models.Post{
			ID:              postID,
			Title:           "Test Title",
			Content:         "Test Content",
			CommentsAllowed: true,
		}, nil)
		mockComments := []models.Comment{
			{
				ID:        "comment1",
				PostID:    postID,
				Text:      "Comment 1",
				ParentID:  "",
				CreatedAt: time.Now(),
			},
			{
				ID:        "comment2",
				PostID:    postID,
				Text:      "Comment 2",
				ParentID:  "comment1",
				CreatedAt: time.Now(),
			},
		}
		mockComment.On("CheckCommentId", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockComment.On("GetComments", mock.Anything, postID, mock.Anything, mock.Anything).
			Return(&mockComments, "endCursor", false, nil)

		resolver := &graph.Resolver{
			Post_:    mockPost,
			Comment_: mockComment,
			Storage:  new(postgres.Storage),
			Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver})))

		var resp struct {
			Post struct {
				ID              string
				Title           string
				Content         string
				CommentsAllowed bool
				Comments        struct {
					Edges []struct {
						Cursor string
						Node   struct {
							ID        string
							PostID    string
							Text      string
							ParentID  *string
							CreatedAt string
						}
					}
					EndCursor   *string
					HasNextPage bool
				}
			}
		}
		query := fmt.Sprintf(`
			query {
				post(id: "%s") {
					id
					title
					content
					commentsAllowed
					comments {
						edges {
							cursor
							node {
								id
								postID
								text
								parentID
								createdAt
							}
						}
						endCursor
						hasNextPage
					}
				}
			}
		`, postID)

		c.MustPost(query, &resp)
		mockPost.AssertExpectations(t)
		mockComment.AssertExpectations(t)
		require.Equal(t, postID, resp.Post.ID)
		require.Equal(t, "Test Title", resp.Post.Title)
		require.Equal(t, "Test Content", resp.Post.Content)
		require.True(t, resp.Post.CommentsAllowed)
		require.Len(t, resp.Post.Comments.Edges, 2)
		require.Equal(t, "comment1", resp.Post.Comments.Edges[0].Node.ID)
		require.Equal(t, "Comment 1", resp.Post.Comments.Edges[0].Node.Text)
		require.Equal(t, "comment2", resp.Post.Comments.Edges[1].Node.ID)
		require.Equal(t, "comment1", *resp.Post.Comments.Edges[1].Node.ParentID)
		require.Equal(t, "Comment 2", resp.Post.Comments.Edges[1].Node.Text)
		require.NotNil(t, resp.Post.Comments.EndCursor)
		require.False(t, resp.Post.Comments.HasNextPage)
	})
}
