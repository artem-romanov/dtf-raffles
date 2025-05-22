package models

import (
	"dtf/game_draw/pkg/dtfapi"
	"errors"
	"fmt"
	"strings"

	"github.com/k3a/html2text"
)

type DataBlock interface {
	Type() string
}

type DataText struct {
	HtmlText string
}

func (d DataText) Type() string { return "text" }

type DataHeader struct {
	Style string
	Text  string
}

func (d DataHeader) Type() string { return "header" }

type Post struct {
	Id     int64
	Title  string
	Text   string // cleaned from html and concatenated text
	Uri    string
	Blocks []DataBlock
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

	var cleanedTextBuilder strings.Builder
	var blocks []DataBlock
	for _, block := range post.Blocks {
		switch b := block.(type) {
		case dtfapi.DataText:
			data := DataText{
				HtmlText: b.HtmlText,
			}

			cleanedText := html2text.HTML2Text(b.HtmlText)
			cleanedTextBuilder.WriteString(cleanedText + "\n")
			blocks = append(blocks, data)
		case dtfapi.DataHeader:
			data := DataHeader{
				Style: b.Style,
				Text:  b.Text,
			}
			cleanedTextBuilder.WriteString(b.Text)
			blocks = append(blocks, data)
		default:
			// do nothing
			continue
		}
	}

	return Post{
		Id:     int64(post.Id),
		Title:  post.Title,
		Uri:    post.Uri,
		Text:   cleanedTextBuilder.String(),
		Blocks: blocks,
	}, nil
}
