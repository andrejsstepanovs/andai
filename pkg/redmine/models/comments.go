package models

type Comment struct {
	Number    int
	UserID    int
	Text      string
	CreatedAt string
}

type Comments []Comment
