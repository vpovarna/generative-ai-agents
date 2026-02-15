package database

import "fmt"

type Document struct {
	Id    string
	Title string
}

func (d *Document) Print() string {
	return fmt.Sprintf("Document_id: %s - Title: %s", d.Id, d.Title)
}

type Chunk struct {
	Id         string
	DocumentID string
	Content    string
	Distance   float64
}
