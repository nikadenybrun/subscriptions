package storage

import "errors"

var (
	ErrNoPostsFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
)
