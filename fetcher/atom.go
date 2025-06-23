package fetcher

import (
	"encoding/xml"
	"html/template"
	"log"
	"news-aggregator/models"
	"news-aggregator/utils"
	"time"
)

type AtomFeed struct {
	ID           string         `xml:"id"`
	Title        AtomText       `xml:"title"`
	Updated      time.Time      `xml:"updated"`
	Authors      []AtomPerson   `xml:"author,omitempty"`
	Links        []AtomLink     `xml:"link"`
	Categories   []AtomCategory `xml:"category,omitempty"`
	Contributors []AtomPerson   `xml:"contributor,omitempty"`
	Generator    *AtomGenerator `xml:"generator,omitempty"`
	Icon         string         `xml:"icon,omitempty"`
	Logo         string         `xml:"logo,omitempty"`
	Rights       AtomText       `xml:"rights,omitempty"`
	Subtitle     AtomText       `xml:"subtitle,omitempty"`
	Entries      []AtomEntry    `xml:"entry"`
}

type AtomEntry struct {
	ID           string         `xml:"id"`
	Title        AtomText       `xml:"title"`
	Updated      string         `xml:"updated"`
	Authors      []AtomPerson   `xml:"author,omitempty"`
	Content      *AtomContent   `xml:"content,omitempty"`
	Links        []AtomLink     `xml:"link"`
	Summary      *AtomText      `xml:"summary,omitempty"`
	Categories   []AtomCategory `xml:"category,omitempty"`
	Contributors []AtomPerson   `xml:"contributor,omitempty"`
	Published    *time.Time     `xml:"published,omitempty"`
	Rights       *AtomText      `xml:"rights,omitempty"`
	Source       *AtomFeed      `xml:"http://www.w3.org/2005/Atom source,omitempty"`
}

type AtomText struct {
	Type string `xml:"type,attr,omitempty"`
	Text string `xml:",chardata"`
}

type AtomPerson struct {
	Name  string `xml:"name"`
	URI   string `xml:"uri,omitempty"`
	Email string `xml:"email,omitempty"`
}

type AtomLink struct {
	Href     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	Hreflang string `xml:"hreflang,attr,omitempty"`
	Title    string `xml:"title,attr,omitempty"`
	Length   string `xml:"length,attr,omitempty"`
}

type AtomCategory struct {
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Label  string `xml:"label,attr,omitempty"`
}

type AtomGenerator struct {
	Text    string `xml:",chardata"`
	URI     string `xml:"uri,attr,omitempty"`
	Version string `xml:"version,attr,omitempty"`
}

type AtomContent struct {
	Type     string `xml:"type,attr,omitempty"`
	Src      string `xml:"src,attr,omitempty"`
	Text     string `xml:",chardata"`
	InnerXML string `xml:",innerxml"`
}

func ParseAtom(data []byte, category string) []models.NewsItem {
	var atom AtomFeed
	if err := xml.Unmarshal(data, &atom); err != nil {
		return nil
	}

	var channelLink string
	for _, link := range atom.Links {
		if link.Rel == "alternate" || link.Rel == "" {
			channelLink = link.Href
			break
		}
	}

	var newsItems []models.NewsItem
	for _, entry := range atom.Entries {
		var itemLink string
		for _, link := range entry.Links {
			if link.Rel == "alternate" || link.Rel == "" {
				itemLink = link.Href
				break
			}
		}

		var dateStr string
		if entry.Published != nil {
			dateStr = entry.Published.Format(time.RFC3339)
		} else if entry.Updated != "" {
			dateStr = entry.Updated
		}

		if dateStr == "" {
			log.Printf("Failed to parse date '' for entry '%s': empty date string", entry.Title.Text)
			continue
		}

		pubTime, err := utils.FormatDate(dateStr)
		if err != nil {
			log.Printf("Failed to parse date '%s' for entry '%s': %v", dateStr, entry.Title.Text, err)
			continue
		}

		var cleanedDescription string
		if entry.Content != nil {
			cleanedDescription = utils.StripHTMLTags(entry.Content.Text)
		} else if entry.Summary != nil {
			cleanedDescription = utils.StripHTMLTags(entry.Summary.Text)
		}

		newsItems = append(newsItems, models.NewsItem{
			Title:        entry.Title.Text,
			Description:  template.HTML(cleanedDescription),
			ChannelLink:  channelLink,
			PubDate:      pubTime,
			Content:      template.HTML(cleanedDescription),
			ItemLink:     itemLink,
			ChannelTitle: atom.Title.Text,
			Category:     category,
			Favicon:      utils.GetFaviconURL(channelLink),
		})
	}

	return newsItems
}
