package models

import "time"

// type Post struct {
// 	ID            int64      `json:"id" pg:"id,pk,autoincrement"`
// 	Title         string     `json:"title"`
// 	Content       string     `json:"content"`
// 	AllowComments bool       `json:"allowComments"`
// 	Comments      []*Comment `json:"comments"`
// }

//	type Comment struct {
//		ID        int64     `json:"id" pg:"id,pk,autoincrement"`
//		Text      string    `json:"text"`
//		Parent    *Comment  `json:"parent,omitempty"`
//		CreatedAt time.Time `json:"createdAt"`
//	}
type Comments struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postID"`
	Text      string    `json:"text"`
	ParentID  string    `json:"parentID"`
	CreatedAt time.Time `json:"createdAt"`
}
type Posts struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	CommentsAllowed bool   `json:"commentsAllowed"`
	// Comments        []Comments `json:"comments,omitempty"`
}
