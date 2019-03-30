package query

import (
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

// @todo to/from JSON
type Query struct {
	Target    Target
	Resource  string
	Scope     string
	Name      string
	Select    []string
	Filter    []string // @todo
	Map       []string // @todo
	RawSource string

	// @todo to force request and ignore cache
	IgnoreCache bool

	// @todo config for querier gets passed on
	// with custom dns or user-agent etc
	Config      map[string]string
	WasPrepared bool

	// ? As a query is prepared, resource queries
	// build their own custom queries which can be stored here
	Parts map[string]interface{}
}

type Queryable interface {
	// Name() string

	Prepare(q *Query)

	Validate(q *Query) error

	// The query executor
	Exec(q *Query) Transaction

	// What can you select from it?
	Selectable() []string

	// Are there nested queryable objects?
	// Queryable() map[string]Queryable

	// What can you configure for the query?
	// Configurable() []string
	// Configure(interface{})

	// What sort of filters are availabe
	// Pipes() [] // for filter/map
}

/*
type AttrValue struct {
    Field string
    Name string
    AsString string
    Value interface{}
}
*/

// @todo
/*
type QueryMultiTarget struct {
	Query
	Targets []Target
}
*/

func (q *Query) AsString() string {
	parts := []string{q.Resource}

	if len(q.Scope) > 0 {
		parts[0] = parts[0] + " " + strings.TrimSpace(q.Scope)
	}

	parts = append(parts, q.Select...)
	return strings.Join(parts, " > ")
}

func (q *Query) TargetFromString(t string) {
	q.Target = TargetFromString(t)
}

func (q *Query) Validate() (bool, []error) {
	if q.Target.RootUrl() == "" {
		return false, []error{ErrorMissingTarget}
	}

	return true, []error{}
}

func (q *Query) SetTarget(t Target) {
	q.Target = t
}

// To group up queries for a target+config to cache/save for later
type Session struct{}

type EntryQueryer struct {
	Executors2   map[string]Queryable
	ErrorCounter int
	Transactions []Transaction
	CacheHits    int
	Logger       *log.Entry
	Cache        *lru.Cache
}

func (t *EntryQueryer) Selectable() []string {
	selectable := []string{}
	for k := range t.Executors2 {
		selectable = append(selectable, k)
	}
	return selectable

}

func (t *EntryQueryer) Queryable() []Queryable {
	items := []Queryable{}

	for _, val := range t.Executors2 {
		if iface, ok := val.(Queryable); ok {
			items = append(items, iface)
		}

	}
	return items
}

func (s *EntryQueryer) Validate(q *Query) error {
	if q.Resource == "" {
		return ErrorMissingResource
	}

	if !q.Target.IsValid() {
		// return ErrorMissingTarget
	}
	return nil
}

// @todo separate network mechanism to queue up requests so double requests are not made
// @todo method for searchall, first, length

func (s *EntryQueryer) Exec(q *Query) Transaction {
	flog := s.Logger.WithFields(log.Fields{"query": q.AsString()})
	flog.Info("Executing query")

	// @todo improve caching
	if !q.IgnoreCache && s.Cache.Contains(q.AsString()) {
		txn, ok := s.Cache.Get(q.AsString())
		if !ok {
			flog.Info("Execute cache miss")
		} else {
			flog.Info("Execute cache hit")
			s.CacheHits += 1
			return txn.(Transaction)
		}
	}

	queryer, ok := s.Executors2[q.Resource]

	// If a executor for the resource is not found
	// check if there is a default executor otherwise give an error
	if !ok {
		return Transaction{
			Query: q,
			Error: ErrorNoExecutor,
		}
	}

	flog.Info("Found queryer for ", q.Resource)

	err := queryer.Validate(q)

	if err != nil {
		return Transaction{
			Query:    q,
			Error:    err,
			Executed: []Queryable{queryer},
		}
	}

	txn := queryer.Exec(q)
	txn.Executed = append(txn.Executed, queryer)
	s.Transactions = append(s.Transactions, txn)
	flog.WithFields(log.Fields{"duration": txn.Duration}).Info("Query duration")
	s.Cache.Add(q.AsString(), txn)
	return txn
}

// @todo the concept of a "session" where queries can be cached for anything in the session window
// @todo pass config to queries
// @todo convert dash-query to underscore
// @todo middlware for query
// @todo check if reading from db/json if not 'live fire'
func Search(q *Query) Transaction {
	searcher := DefaultQueryer()
	err := searcher.Validate(q)

	if err != nil {
		return Transaction{Error: err}
	}
	return searcher.Exec(q)
}

func DefaultQueryer() *EntryQueryer {
	logger := log.WithFields(log.Fields{
		"file":   "query/query.go",
		"struct": "Query",
	})
	var cache, _ = lru.New(4096)

	return &EntryQueryer{
		Executors2: map[string]Queryable{
			"http": DefaultHttpQueryer(),
			"ssl":  DefaultSslQueryer(),
			"dns":  DefaultDnsQueryer(),
		},
		Logger: logger,
		Cache:  cache,
	}

}

type Transaction struct {
	Id        string
	Query     *Query
	Results   []QueryResult
	Duration  time.Duration
	FromCache bool
	Error     error
	Executed  []Queryable
}

func (t *Transaction) GetErrors() error {
	return t.Error
}

func (t *Transaction) GetResults() []QueryResult {
	return t.Results
}

/*
func QueryJsonCache(q *Query) Transaction {
	queryer := DefaultHttpEntryQueryer()
	return queryer.Query(q)
}
*/
