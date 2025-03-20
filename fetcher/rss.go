package fetcher

import (
	"encoding/xml"
	"html/template"
	"log"
	"news-aggregator/models"
	"news-aggregator/utils"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title           string    `xml:"title"`
	Description     string    `xml:"description"`
	Link            string    `xml:"link"`
	Generator       string    `xml:"generator"`
	LastBuildDate   string    `xml:"lastBuildDate"`
	Language        string    `xml:"language"`
	UpdatePeriod    string    `xml:"updatePeriod"`
	UpdateBase      string    `xml:"updateBase"`
	UpdateFrequency int       `xml:"updateFrequency"`
	Image           Image     `xml:"image"`
	Items           []RSSItem `xml:"item"`
}

type Image struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	URL         string `xml:"url"`
	Description string `xml:"description"`
}

type RSSItem struct {
	Title       string        `xml:"title"`
	Description string        `xml:"description"`
	Link        string        `xml:"link"`
	Content     template.HTML `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	PubDate     string        `xml:"pubDate"`
	Media       Media         `xml:"media:content"`
	Creator     string        `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Comments    string        `xml:"wfw:comment"`
	Guid        string        `xml:"guid"`
}

type Media struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

func ParseRSS(body []byte) []models.NewsItem {
	var rss RSS
	err := xml.Unmarshal(body, &rss)
	if err != nil {
		log.Println("Error parsing RSS feed:", err)
		return nil
	}

	var items []models.NewsItem
	for _, item := range rss.Channel.Items {
		cleanedDescription := utils.StripHTMLTags(string(item.Description))
		items = append(items, models.NewsItem{
			Title:        item.Title,
			Description:  template.HTML(cleanedDescription),
			Link:         item.Link,
			PubDate:      utils.FormatDate(item.PubDate),
			Content:      item.Content,
			MediaURL:     item.Media.URL,
			Creator:      item.Creator,
			Comments:     item.Comments,
			Guid:         item.Guid,
			ChannelLink:  rss.Channel.Link,
			ChannelTitle: rss.Channel.Title,
		})
	}

	return items
}
