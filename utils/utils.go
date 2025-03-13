package utils

import (
	"regexp"
	"strings"
	"time"
)

func StripHTMLTags(html string) string {
	var result strings.Builder
	var inTag bool

	for _, char := range html {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}
	return result.String()
}

func stripHTMLTagsReg(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

func FormatDate(dateStr string) string {
	t, err := time.Parse(time.RFC1123Z, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("02.01.2006 15:04")
}

func GetCurrentDate() string {
	return time.Now().Format("02.01.2006")
}
