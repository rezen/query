package query

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rezen/query/requests"
	// "github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type DocQuery struct {
	Target *Target
	Select []string
	Txn    *requests.Transaction
	Doc    *goquery.Document
	// @todo config for querier gets passed on
	// with custom dns or user-agent etc
	start     time.Time
	Duration  time.Duration
	Logger    *log.Entry
	Executors map[string]docQueryExecutor
}

func DefaultDocQueryer() DocQuery {
	return DocQuery{
		Logger: log.WithFields(log.Fields{
			"struct": "DocQuery",
		}),
		Executors: map[string]docQueryExecutor{
			"words":    executeQueryWords,
			"comments": executeQueryComments,
			"title":    executeQueryTitle,
			"default":  executeQuerySelector,
		},
	}
}

type docQueryExecutor func(q *DocQuery) ([]QueryResult, error)

func (d *DocQuery) Selectable() []string {
	items := []string{}
	for k := range d.Executors {
		items = append(items, k)
	}
	return items
}

func (q *DocQuery) Attr(attr string) string {
	// @todo attr parser with index?
	switch attr {
	case "title":
		el := q.Doc.Find("title").First()
		return el.Text()
	}
	return ""
}

func (q *DocQuery) Prepare() {
	var err error
	reader := strings.NewReader(q.Txn.Body())
	q.Doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		q.Logger.WithFields(log.Fields{"source": "goquery"}).Error(err)
	}
	q.start = time.Now()
}

func (q *DocQuery) Execute() ([]QueryResult, error) {
	flog := q.Logger.WithFields(log.Fields{"select": q.Select})
	flog.Info("Querying doc")

	// @todo a way to return DocResult
	if len(q.Select) == 0 {
		return []QueryResult{}, ErrorMissingSelector
	}

	branch := q.Select[0]
	defer func() {
		q.Duration = time.Since(q.start)
	}()

	executor, exists := q.Executors[branch]

	if !exists {
		executor = q.Executors["default"]
	}
	return executor(q)
}

// @todo force attrs when encoding to json or map
// @todo concept of dynamic attributes?
type ElementResult struct {
	Target   *Target
	El       *goquery.Selection
	NodeName string `json:"type"`
	Selector string `json:"selector"`
	Idx      int    `json:"idx"`
}

func (q ElementResult) HasAttr(attr string) bool {
	_, exists := q.El.Attr(attr)
	return exists
}

func (q ElementResult) AsText() string {
	raw, _ := goquery.OuterHtml(q.El)
	return raw

}

func (q ElementResult) Attr(attr string) string {
	switch attr {
	case "type":
		return q.NodeName
	case "innertext":
		return q.El.Text()
	default:
		val, exists := q.El.Attr(attr)

		if !exists {
			val = "" // @todo
		}
		return val
	}
}

func executeQueryWords(q *DocQuery) ([]QueryResult, error) {
	results := []QueryResult{}
	words := DocWords(*q.Doc)

	for _, word := range words {
		results = append(results, &TextResult{"words", word})
	}
	return results, nil
}

func executeQueryComments(q *DocQuery) ([]QueryResult, error) {
	results := []QueryResult{}
	comments := DocComments(*q.Doc)
	for _, comment := range comments {
		results = append(results, &TextResult{"comments", comment})
	}
	return results, nil
}

func executeQueryTitle(q *DocQuery) ([]QueryResult, error) {
	el := q.Doc.Find("title").First()
	return []QueryResult{&TextResult{"title", el.Text()}}, nil
}

func executeQuerySelector(q *DocQuery) ([]QueryResult, error) {
	// @todo update selector, replacing special terms with css selectors
	// for example `css` translates to `link[rel=stylesheet]`
	// or links-external to a:not([href^='{{target}}'])
	selector := strings.Join(q.Select, " ")
	findings := []QueryResult{}
	q.Doc.Find(selector).Each(func(i int, el *goquery.Selection) {
		findings = append(findings, ElementResult{
			Target:   q.Target,
			El:       el,
			Selector: selector,
			Idx:      i,
			NodeName: goquery.NodeName(el),
		})
	})
	return findings, nil
}

/*		return
		case "retirejs":
		// repo := retirejs.GetRepository()
		// findings := repo.CheckUri([]byte(source))
		// vulns := retirejs.EvaluateFindings(findings)
*/
