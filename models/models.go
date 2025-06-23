package models

import (
	"html/template"
	"time"
)

type NewsItem struct {
	Title        string        `json:"title"`
	Description  template.HTML `json:"description"`
	ChannelLink  string        `json:"channelLink"`
	PubDate      time.Time     `json:"pubDate"`
	Content      template.HTML `json:"content"`
	MediaURL     string        `json:"mediaURL"`
	Creator      string        `json:"creator"`
	Comments     string        `json:"comments"`
	Guid         string        `json:"guid"`
	ItemLink     string        `json:"itemLink"`
	ChannelTitle string        `json:"channelTitle"`
	Category     string        `json:"category"`
	Favicon      string        `json:"favicon"`
}
