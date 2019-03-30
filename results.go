package query

import (
	"math/rand"
	"time"
)

type QueryResult interface {
	Attr(attr string) string
	HasAttr(attr string) bool
	AsText() string
}

type ResultWrap struct {
	Real interface{}
}

func (r *ResultWrap) Attr(attr string) string {
	return ""
}

func (r *ResultWrap) HasAttr(attr string) bool {
	return HasAttr(r.Real, attr)
}

func (r *ResultWrap) AsText() string {
	return AsText(r.Real)
}

func ForceResult(v interface{}) QueryResult {
	return &ResultWrap{v}
}

func ResultsMap(results []QueryResult, fn func(QueryResult) QueryResult) []QueryResult {
	mapped := []QueryResult{}
	for _, result := range results {
		mapped = append(mapped, fn(result))
	}
	return mapped
}

func ResultsFilter(results []QueryResult, fn func(QueryResult) bool) []QueryResult {
	filtered := []QueryResult{}
	for _, result := range results {
		if fn(result) {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func ResultsRandom(results []QueryResult) QueryResult {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	idx := r.Intn(len(results))
	return results[idx]
}

func ResultsUnique(results []QueryResult) []QueryResult {
	return results
}

// func (r *QueryResult) First(attr)
// func (r *QueryResult) Filter(attr)

type NoResult struct{}

func (r *NoResult) Attr(attr string) string {
	return ""
}

func (r *NoResult) HasAttr(attr string) bool {
	return false
}

func (r *NoResult) AsText() string {
	return ""
}

type TextResult struct {
	Key  string
	Text string
}

func (r TextResult) Attr(attr string) string {
	return r.Text
}

func (r TextResult) HasAttr(attr string) bool {
	return false
}

func (r TextResult) AsText() string {
	return r.Text
}

type MapResult struct {
	Data map[string]string
}

func (r *MapResult) Attr(attr string) string {
	return r.Data[attr]
}

func (r *MapResult) HasAttr(attr string) bool {
	_, exist := r.Data[attr]
	return exist
}

func (r *MapResult) AsText() string {
	// @todo
	text := ""
	for key, value := range r.Data {
		text += " - " + key + ": " + value + "\n"
	}
	return text

}
