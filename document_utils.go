package query

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
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

// https://stackoverflow.com/questions/44441665/how-to-extract-only-text-from-html-in-golang
// https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
func HtmlWords(body io.Reader) []string {
	z := html.NewTokenizer(body)
	var words []string

	var lastToken html.Token
	aggregate := ""
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			regex := regexp.MustCompile(`\s+`)
			aggregate = regex.ReplaceAllString(aggregate, " ")

			for _, w := range strings.Split(aggregate, " ") {

				w = strings.Trim(w, ",.;!'?:()& “”\n")
				if len(w) > 1 {
					words = append(words, w)
				}
			}

			return uniqueWords(words)
		case tt == html.StartTagToken:
			t := z.Token()
			lastToken = t

			for _, a := range t.Attr {
				if a.Key == "title" || a.Key == "alt" {

					if len(a.Val) == 0 {
						continue
					}

					// HTML in title ?
					if string(a.Val[0]) != "<" {
						aggregate += " " + a.Val
					}
				}
			}

		case tt == html.TextToken:
			if lastToken.Data == "script" || lastToken.Data == "noscript" || lastToken.Data == "style" {
				continue
			}

			aggregate += " " + string(z.Text())
		}
	}

	return []string{}
}

func uniqueWords(words []string) []string {
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
