package dtfapi

import (
	"time"

	"github.com/guregu/null/v6"
)

type Tokens struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiration time.Time
}

func (t Tokens) IsAccessValid() bool {
	if t.AccessToken == "" {
		return false
	}

	diff := time.Until(t.AccessExpiration)
	if diff.Microseconds() <= 0 {
		return false
	}

	return true
}

// POST Structs

type DataBlock interface {
	Type() string
}

type DataText struct {
	HtmlText string
}

func (dt DataText) Type() string {
	return "text"
}

type DataHeader struct {
	Style string
	Text  string
}

func (dt DataHeader) Type() string {
	return "header"
}

// TODO: add other types in future,
// right now it's fine to have only Text block
// type DataImage struct {}

type BlogPost struct {
	Id        int
	Title     string
	Uri       string
	Blocks    []DataBlock
	RepliedTo null.Int32 // if not null - it is a reply to that original post
}

// USER Structs
type UserInfo struct {
	Id   int
	Url  string
	Name string
}
