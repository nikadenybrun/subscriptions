package models

import "time"

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postID"`
	Text      string    `json:"text"`
	ParentID  string    `json:"parentID"`
	CreatedAt time.Time `json:"createdAt"`
}
type Post struct {
	ID              string
	Title           string
	Content         string
	CommentsAllowed bool
	CreatedAt       time.Time
}
