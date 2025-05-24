package inmemory

import (
	"context"
	"fmt"
	"sort"
	"subscriptions/internal/config"
	models "subscriptions/internal/domain"
	"subscriptions/internal/graphql/graph/model"
	"sync"

	"github.com/google/uuid"
)

type InMemoryStorage struct {
	mu       sync.RWMutex
	posts    map[string]*models.Post
	comments map[string]*models.Comment
}

func NewPost() *InMemoryStorage {
	return &InMemoryStorage{
		posts:    make(map[string]*models.Post),
		comments: make(map[string]*models.Comment),
	}
}

func NewStorage(storagePath string, cfg config.DBSaver) (*InMemoryStorage, error) {
	return &InMemoryStorage{
		posts:    make(map[string]*models.Post),
		comments: make(map[string]*models.Comment),
	}, nil
}
func (s *InMemoryStorage) CloseDB() error {
	return nil
}

func (s *InMemoryStorage) SavePost(ctx context.Context, p *model.Post) (string, error) {
	// op := "In-memory.SavePost"
	s.mu.Lock()
	defer s.mu.Unlock()
	p.ID = uuid.New().String()
	var post models.Post
	post.ID, post.Title, post.Content, post.CommentsAllowed, post.CreatedAt = p.ID, p.Title, p.Content, p.CommentsAllowed, p.CreatedAt
	s.posts[p.ID] = &post
	return p.ID, nil
}

func (s *InMemoryStorage) GetPost(ctx context.Context, id string) (*models.Post, error) {
	op := "In-memory.GetPost"
	s.mu.RLock()
	defer s.mu.RUnlock()
	post, ok := s.posts[id]
	if !ok {
		return nil, fmt.Errorf("post not found", op)
	}
	return post, nil
}

func (s *InMemoryStorage) GetAll(ctx context.Context) (*[]models.Post, error) {
	// op := "In-memory.GetAll"
	s.mu.RLock()
	defer s.mu.RUnlock()
	posts := make([]models.Post, 0, len(s.posts))
	for _, p := range s.posts {
		posts = append(posts, *p)
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.Before(posts[j].CreatedAt)
	})
	return &posts, nil
}
func (s *InMemoryStorage) SaveComment(ctx context.Context, c *model.Comment) (string, error) {
	// op := "In-memory.SaveComment"
	s.mu.Lock()
	defer s.mu.Unlock()

	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	comment := models.Comment{
		ID:        c.ID,
		PostID:    c.PostID,
		Text:      c.Text,
		CreatedAt: c.CreatedAt,
	}
	if c.ParentID != nil {
		comment.ParentID = *c.ParentID
	}
	s.comments[c.ID] = &comment
	return c.ID, nil
}

func (s *InMemoryStorage) GetComments(ctx context.Context, postId string, first *int32, after *string) (*[]models.Comment, string, bool, error) {
	op := "In-memory.GetComments"
	s.mu.RLock()
	defer s.mu.RUnlock()
	if first == nil {
		return nil, "", false, fmt.Errorf("You need to add param first: %w")
	}

	var filtered []models.Comment
	if *first == 0 {
		return &filtered, "", true, nil
	}
	for _, c := range s.comments {
		if c.PostID == postId {
			filtered = append(filtered, *c)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
	})

	startIndex := 0
	if after != nil && *after != "" {
		for i, c := range filtered {
			if c.ID == *after {
				startIndex = i + 1
				break
			}
		}
		if startIndex == 0 {
			return nil, "", false, fmt.Errorf("invalid cursor: %w", op)
		}
	}

	limit := len(filtered) - startIndex
	if int(*first) < limit {
		limit = int(*first)
	}

	paginated := filtered[startIndex : startIndex+limit]

	var endCursor string
	if len(paginated) > 0 {
		endCursor = paginated[len(paginated)-1].ID
	}

	hasNextPage := (startIndex + limit) < len(filtered)

	return &paginated, endCursor, hasNextPage, nil
}

func (s *InMemoryStorage) CheckCommentId(ctx context.Context, comtId *string, postId string) error {
	op := "In-memory.CheckCommentId"
	s.mu.RLock()
	defer s.mu.RUnlock()
	if comtId == nil {
		return fmt.Errorf("comment not found", op)
	}
	comment, ok := s.comments[*comtId]
	if !ok {
		return fmt.Errorf("comment not found", op)
	}
	if comment.PostID != postId {
		return fmt.Errorf("comment does not belong to the specified post", op)
	}

	return nil
}
