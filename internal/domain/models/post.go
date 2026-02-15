package models

import (
	"dtf/game_draw/pkg/dtfapi"
	"errors"
	"fmt"
	"strings"

	"github.com/guregu/null/v6"
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

type DataList struct {
	List []string
}

func (dl DataList) Type() string { return "list" }

// String() returns cleaned from HTML text.
// All html tags will be removed.
func (dl DataList) String() string {
	if len(dl.List) == 0 {
		return ""
	}

	var b strings.Builder
	for i, item := range dl.List {
		if i == 0 {
			b.WriteByte('\n')
		}
		cleandText := html2text.HTML2Text(item)
		b.WriteString("- ")
		b.WriteString(cleandText)
		b.WriteByte('\n')
	}

	return b.String()
}

type Post struct {
	Id        int64
	Title     string
	Text      string // cleaned from html and concatenated text
	Uri       string
	Blocks    []DataBlock
	RepliedTo null.Int32 // if not null - post is a reply to that post
}

func (p Post) Print() {
	fmt.Println("_________")
	fmt.Printf("Blog Post #%d\n", p.Id)
	fmt.Printf("Title: %s\n", p.Title)
	fmt.Printf("URI: %s\n", p.Uri)
}

func FromDtfPost(post dtfapi.BlogPost) (Post, error) {
	if post.Id <= 0 {
		return Post{}, errors.New("can't map. Id is less than 1")
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

		case dtfapi.DataList:
			data := DataList{
				List: b.Items(),
			}
			cleanedText := data.String()
			cleanedTextBuilder.WriteString(cleanedText)
			blocks = append(blocks, data)

		default:
			// do nothing
			continue
		}
	}

	return Post{
		Id:        int64(post.Id),
		Title:     post.Title,
		Uri:       post.Uri,
		Text:      cleanedTextBuilder.String(),
		Blocks:    blocks,
		RepliedTo: post.RepliedTo,
	}, nil
}
