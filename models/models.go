package models

import (
	"html/template"
	"time"
)

type NewsItem struct {
	Title        string        `json:"title"`
	Description  template.HTML `json:"description"`
	Link         string        `json:"link"`
	PubDate      time.Time     `json:"pubDate"`
	Content      template.HTML `json:"content"`
	MediaURL     string        `json:"mediaURL"`
	Creator      string        `json:"creator"`
	Comments     string        `json:"comments"`
	Guid         string        `json:"guid"`
	ChannelLink  string        `json:"channelLink"`
	ChannelTitle string        `json:"channelTitle"`
	Category     string        `json:"category"`
}
