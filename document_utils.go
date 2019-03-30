package query

import (
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"sort"
	"strings"
)

func findCommentNodes(dst, src *goquery.Selection) *goquery.Selection {
	src.Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#comment" {
			dst = dst.AddSelection(s)
		} else {
			dst = findCommentNodes(dst, s)
		}
	})
	return dst
}

// @todo improve
func DocComments(doc goquery.Document) []string {
	comments := []string{}
	commentNodes := new(goquery.Selection)
	root := doc.Find("*")
	commentNodes = findCommentNodes(commentNodes, root)
	commentNodes.Each(func(i int, s *goquery.Selection) {
		h, err := goquery.OuterHtml(s)
		if err != nil {
			panic(err)
		}

		comments = append(comments, h)
	})

	return comments
}

func DocWords(doc goquery.Document) []string {
	aggregate := ""
	doc.Find("script,style").Remove()
	// .RemoveFiltered("script")
	doc.Find("a,img,button").Each(func(i int, el *goquery.Selection) {
		name := goquery.NodeName(el)

		if name == "script" {
			return
		}

		if name == "iframe" {
			return
		}

		for _, attr := range []string{"title", "alt"} {
			val, exists := el.Attr(attr)
			if exists && len(val) > 0 {
				aggregate += " " + strings.TrimSpace(val)
			}
		}
	})

	doc.Find("html").Each(func(i int, el *goquery.Selection) {
		regex := regexp.MustCompile(`[,\.!?)&*#]`)
		text := el.Text()
		text = regex.ReplaceAllString(text, " ")
		aggregate += " " + strings.TrimSpace(text)
	})

	regex := regexp.MustCompile(`\s+`)
	aggregate = regex.ReplaceAllString(aggregate, " ")
	words := strings.Fields(aggregate)

	uniqueWords := []string{}
	wordMap := map[string]bool{}

	for _, word := range words {
		// @todo clean words
		if _, ok := wordMap[word]; ok {
			continue
		}

		wordMap[word] = true
		uniqueWords = append(uniqueWords, word)
	}
	sort.Strings(uniqueWords)

	return uniqueWords
}
