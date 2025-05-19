package models

import (
	"dtf/game_draw/pkg/dtfapi"
	"errors"
	"fmt"
)

type Post struct {
	Id    int64
	Title string
	Uri   string
}

func (p Post) Print() {
	fmt.Println("_________")
	fmt.Printf("Blog Post #%d\n", p.Id)
	fmt.Printf("Title: %s\n", p.Title)
	fmt.Printf("URI: %s\n", p.Uri)
}

func FromDtfPost(post dtfapi.BlogPost) (Post, error) {
	if post.Id <= 0 {
		return Post{}, errors.New("Can't map. Id is less than 1")
	}

	return Post{
		Id:    int64(post.Id),
		Title: post.Title,
		Uri:   post.Uri,
	}, nil
}
