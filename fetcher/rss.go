package fetcher

import (
	"encoding/xml"
	"html/template"
	"io"
	"log"
	"news-aggregator/models"
	"news-aggregator/utils"
	"strings"
	"time"
)

type RSS struct {
	Version        string            `xml:"version,attr,omitempty"`
	Title          string            `xml:"channel>title,omitempty"`
	Link           string            `xml:"channel>link,omitempty"`
	Description    string            `xml:"channel>description"`
	Language       string            `xml:"channel>language,omitempty"`
	Copyright      string            `xml:"channel>copyright,omitempty"`
	ManagingEditor string            `xml:"channel>managingEditor,omitempty"`
	WebMaster      string            `xml:"channel>webMaster,omitempty"`
	PubDate        string            `xml:"channel>pubDate,omitempty"`
	LastBuildDate  string            `xml:"channel>lastBuildDate,omitempty"`
	Category       string            `xml:"channel>category,omitempty"`
	Generator      string            `xml:"channel>generator,omitempty"`
	Docs           string            `xml:"channel>docs,omitempty"`
	Cloud          *Cloud            `xml:"channel>cloud,omitempty"`
	TTL            int               `xml:"channel>ttl,omitempty"`
	Image          *Image            `xml:"channel>image,omitempty"`
	Rating         string            `xml:"channel>rating,omitempty"`
	TextInput      *TextInput        `xml:"channel>textInput,omitempty"`
	SkipHours      *SkipHours        `xml:"channel>skipHours,omitempty"`
	SkipDays       *SkipDays         `xml:"channel>skipDays,omitempty"`
	Items          []Item            `xml:"channel>item"`
	RDFItems       []Item            `xml:"item"`
	Namespaces     map[string]string `xml:"-"`
}

type Cloud struct {
	Domain            string `xml:"domain,attr"`
	Port              int    `xml:"port,attr"`
	Path              string `xml:"path,attr"`
	RegisterProcedure string `xml:"registerProcedure,attr"`
	Protocol          string `xml:"protocol,attr"`
}

type Image struct {
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Width       int    `xml:"width,omitempty"`
	Height      int    `xml:"height,omitempty"`
	Description string `xml:"description,omitempty"`
}

type TextInput struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Name        string `xml:"name"`
	Link        string `xml:"link"`
}

type SkipHours struct {
	Hours []int `xml:"hour"`
}

type SkipDays struct {
	Days []string `xml:"day"`
}

type Item struct {
	Title       string `xml:"title,omitempty"`
	Link        string `xml:"link,omitempty"`
	Description string `xml:"description,omitempty"`

	Author    string     `xml:"author,omitempty"`
	Category  []string   `xml:"category,omitempty"`
	Comments  string     `xml:"comments,omitempty"`
	Enclosure *Enclosure `xml:"enclosure,omitempty"`
	GUID      *GUID      `xml:"guid,omitempty"`
	PubDate   string     `xml:"pubDate,omitempty"`
	Source    *Source    `xml:"source,omitempty"`

	About   string      `xml:"about,attr,omitempty"`
	DC      *DublinCore `xml:"http://purl.org/dc/elements/1.1/ dc,omitempty"`
	Content *Content    `xml:"http://purl.org/rss/1.0/modules/content/ encoded,omitempty"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int    `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type GUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr,omitempty"`
}

type Source struct {
	Value string `xml:",chardata"`
	URL   string `xml:"url,attr"`
}

type DublinCore struct {
	Creator     string    `xml:"creator,omitempty"`
	Date        time.Time `xml:"date,omitempty"`
	Description string    `xml:"description,omitempty"`
	Language    string    `xml:"language,omitempty"`
	Publisher   string    `xml:"publisher,omitempty"`
	Subject     string    `xml:"subject,omitempty"`
	Title       string    `xml:"title,omitempty"`
}

type Content struct {
	Encoded string `xml:",chardata"`
}

func ParseRSS(data []byte) []models.NewsItem {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil
	}
	// log.Printf("Channel Title: %s", rss.Title)
	// log.Printf("Channel Link: %s", rss.Link)
	// log.Printf("Channel Description: %s", rss.Description)
	if rss.Title == "" {
		rss.Title = "Title is missing"
	}
	if rss.Link == "" {
		rss.Link = ExtractLink(data)
		log.Println("Main link field empty or not find, trying alternative parsing -", rss.Link)

	}

	var newsItems []models.NewsItem
	for _, item := range rss.Items {
		cleanedDescription := utils.StripHTMLTags(string(item.Description))
		pubTime, err := utils.FormatDate(item.PubDate)
		if err != nil {
			log.Printf("Failed to parse date '%s' for item '%s': %v",
				item.PubDate, item.Title, err)
			continue
		}
		newsItems = append(newsItems, models.NewsItem{
			Title:        item.Title,
			Description:  template.HTML(cleanedDescription),
			Link:         item.Link,
			PubDate:      pubTime,
			Creator:      item.Author,
			Comments:     item.Comments,
			Guid:         item.GUID.Value,
			ChannelLink:  rss.Link,
			ChannelTitle: rss.Title,
		})
	}

	return newsItems
}
func ExtractLink(data []byte) string {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error decoding XML: %v", err)
			}
			break
		}
		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == "link" {
				var link string
				if err := decoder.DecodeElement(&link, &elem); err == nil {
					return strings.TrimSpace(link)
				}
			}
		}
	}
	return ""
}
