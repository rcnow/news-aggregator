package fetcher

import (
	"encoding/xml"
	"html/template"
	"log"
	"news-aggregator/models"
	"news-aggregator/utils"
)

type AtomFeed struct {
	XMLName  xml.Name    `xml:"feed"`
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle"`
	Links    []AtomLink  `xml:"link"`
	Entries  []AtomEntry `xml:"entry"`
}

type AtomEntry struct {
	Title   string        `xml:"title"`
	Content template.HTML `xml:"content"`
	Link    AtomLink      `xml:"link"`
	Updated string        `xml:"updated"`
}

type AtomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func ParseAtom(body []byte) []models.NewsItem {
	var atom AtomFeed
	err := xml.Unmarshal(body, &atom)
	if err != nil {
		log.Println("Error parsing Atom feed:", err)
		return nil
	}
	var channelLink string
	for _, link := range atom.Links {
		if link.Rel == "alternate" {
			channelLink = link.Href
			break
		}
	}
	var items []models.NewsItem
	for _, entry := range atom.Entries {
		pubTime, err := utils.FormatDate(entry.Updated)
		if err != nil {
			log.Printf("Failed to parse date '%s' for entry '%s': %v",
				entry.Updated, entry.Title, err)
			continue
		}

		cleanedDescription := utils.StripHTMLTags(string(entry.Content))
		items = append(items, models.NewsItem{
			Title:        entry.Title,
			Description:  template.HTML(cleanedDescription),
			Link:         entry.Link.Href,
			PubDate:      pubTime,
			Content:      entry.Content,
			ChannelLink:  channelLink,
			ChannelTitle: atom.Title,
		})
	}

	return items
}
