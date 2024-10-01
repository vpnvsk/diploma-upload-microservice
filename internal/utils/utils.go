package utils

import "regexp"

func RemoveHTMLTags(text string) string {
	re := regexp.MustCompile(`<.*?>`)
	cleanText := re.ReplaceAllString(text, " ")
	return cleanText
}
