package query

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rezen/query/http"
	log "github.com/sirupsen/logrus"

	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func QueryHttp(q *Query) Transaction {
	queryer := DefaultHttpQueryer()
	return queryer.Query(q)
}

type HttpQuery struct {
	Target            *Target
	Path              string
	Name              string
	Select            []string
	Headers           map[string]string
	UseAuthentication bool
	Username          string
	Password          string
	FollowRedirects   bool
	Timeout           int
	IgnoreBadSSL      bool
	Txn               *http.Transaction
	Method            string
	Data              string
}

func (q *HttpQuery) Configure(config map[string]string) {
	// @todo each Queryable should also hook into configure
	if len(config) == 0 {
		return
	}

	if val, ok := config["username"]; ok {
		q.Username = val
	}

	if val, ok := config["password"]; ok {
		q.Username = val
	}

	if _, ok := config["ignore_ssl"]; ok {
		q.IgnoreBadSSL = true
	}

	if _, ok := config["timeout"]; ok {
		// q.Timeout // @todo parse int
	}

}

type HttpQueryer struct {
	// @todo executors should have middlware notion
	Executors  map[string]queryExecutor
	LastError  error
	ErrorCount int
	Logger     *log.Entry
	Cache      *lru.Cache
}

func (q *HttpQueryer) Validate(query *Query) error {
	return nil
}

func (q *HttpQueryer) Query(query *Query) Transaction {

	start := time.Now()
	qry := &HttpQuery{
		Target: &query.Target,
		Path:   query.Scope,
		Select: query.Select,
	}

	qry.Configure(query.Config)
	// @todo prepare should decorate query as well as get executor?
	qry = q.Prep(qry)
	query.WasPrepared = true
	results, err := q.Execute(qry)
	return Transaction{
		Query:    query,
		Results:  results,
		Duration: time.Since(start),
		Error:    err,
	}

}
func (q *HttpQueryer) Prep(query *HttpQuery) *HttpQuery {
	// @todo cache http request
	requestor := http.CreateRequestorWithTarget(query.Target.Url)
	query.Txn = requestor.Get(query.Path)

	//
	// @todo error handling

	return query
}

func (q *HttpQueryer) Prepare(query *Query) {

}

func (q *HttpQueryer) Exec(query *Query) Transaction {
	return q.Query(query)
}

func (q *HttpQueryer) Execute(query *HttpQuery) ([]QueryResult, error) {
	var results []QueryResult
	var err error

	if len(query.Select) == 0 {
		results = []QueryResult{}
		err = errors.New("Missing selector for http")
	} else {
		branch := query.Select[0]
		executor, exists := q.Executors[branch]
		if !exists {
			results = []QueryResult{}
			err = ErrorInvalidSelector
		} else {
			results, err = executor(query)
		}

		q.LastError = err
	}

	if err != nil {
		q.ErrorCount += 1
	}
	return results, err
}

func (q *HttpQueryer) Selectable() []string {
	items := []string{}
	for k := range q.Executors {
		items = append(items, k)
	}
	sort.Strings(items)
	return items
}

/*
// @todo explore?
type Attribute struct {
	Name string
	Queryer func ...
	Calculated bool // Is field calculated, like regex?
	IncludeInAll bool // Include in * query data
	Nested bool // Can be queried ?
}
*/
func DefaultHttpQueryer() *HttpQueryer {
	// When exporting to JSON, iterate through executors
	// Check environment for requirements before
	// making queryer available
	executors := map[string]queryExecutor{
		// "?" // For help
		// "cookie" to parse out individual cookies
		"*":            queryAll,
		"body":         queryBody,
		"header":       queryHeader,
		"hash":         queryHash,
		"sha256":       querySha1,
		"doc":          queryDoc,
		"status-code":  queryStatusCode,
		"ip":           queryIp,
		"regex":        queryRegex,
		"redirects":    queryRedirects,
		"virustotal":   queryVirustotal,
		"safebrowsing": querySafeBrowsing,
		// "browser": queryBrowser, @todo for real browser execution
		// "seo" // Check keyboard richness
	}
	var cache, _ = lru.New(4096)

	return &HttpQueryer{
		Executors:  executors,
		LastError:  nil,
		ErrorCount: 0,
		Logger: log.WithFields(log.Fields{
			"file":   "query/http.go",
			"struct": "HttpQueryer",
		}),
		Cache: cache,
	}
}

type queryExecutor func(q *HttpQuery) ([]QueryResult, error)

func queryDoc(q *HttpQuery) ([]QueryResult, error) {
	selector := q.Select[1:]
	query := DefaultDocQueryer()
	query.Target = q.Target
	query.Select = selector
	query.Txn = q.Txn
	query.Prepare()
	return query.Execute()
}

func queryAll(q *HttpQuery) ([]QueryResult, error) {
	results := []QueryResult{}
	include := []queryExecutor{
		queryStatusCode,
		queryIp,
		queryRedirects,
		queryHash,
		queryHeader,
	}

	for _, fn := range include {
		res, err := fn(q)
		if err != nil {
			continue
		}

		results = append(results, res...)
	}

	return results, nil
}
func queryStatusCode(q *HttpQuery) ([]QueryResult, error) {
	value := strconv.Itoa(q.Txn.StatusCode())
	return []QueryResult{&TextResult{"status_code", value}}, nil
}

func queryIp(q *HttpQuery) ([]QueryResult, error) {
	return []QueryResult{&TextResult{"ip", q.Txn.IP}}, nil
}

func queryRegex(q *HttpQuery) ([]QueryResult, error) {
	pattern := q.Select[1]
	body := q.Txn.Body()
	rex, err := regexp.Compile(pattern)

	// @todo regex all vs first
	if err != nil {
		return []QueryResult{}, err
	}

	matches := rex.FindAllString(body, -1)
	results := []QueryResult{}

	for _, match := range matches {
		results = append(results, &TextResult{"regex", match})
	}
	return results, nil
}

// @todo should affect prepare
func queryRedirects(q *HttpQuery) ([]QueryResult, error) {
	return []QueryResult{&TextResult{"redirects", strconv.Itoa(q.Txn.Redirects)}}, nil
}

func queryBody(q *HttpQuery) ([]QueryResult, error) {
	return []QueryResult{&TextResult{"body", q.Txn.Body()}}, nil
}

func querySha1(q *HttpQuery) ([]QueryResult, error) {
	h := sha256.New()
	h.Write([]byte(q.Txn.Body()))
	hash := hex.EncodeToString(h.Sum(nil))
	return []QueryResult{&TextResult{"sha256", "sha256:" + hash}}, nil
}

func queryHash(q *HttpQuery) ([]QueryResult, error) {
	return []QueryResult{&TextResult{"hash", q.Txn.BodyHash()}}, nil
}

func queryHeader(q *HttpQuery) ([]QueryResult, error) {
	// If no specific header is selected
	if len(q.Select) == 1 {
		headers := q.Txn.Headers()
		headers["status-code"] = strconv.Itoa(q.Txn.StatusCode())
		return []QueryResult{&MapResult{headers, "headers"}}, nil
	}

	// also option for key patterns such as x-*
	// ... or if a specific header is selected
	var value string
	header := q.Select[1]
	header = strings.ToLower(header)

	if header == "status-code" {
		value = strconv.Itoa(q.Txn.StatusCode())
	} else if q.Txn.HeaderExists(header) {
		value = q.Txn.Header(header)
	}

	return []QueryResult{&TextResult{header, value}}, nil
}
