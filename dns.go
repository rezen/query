package query

import (
	"github.com/rezen/query/requests/dns"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type DnsQueryer struct {
	Executors  map[string]dnsQueryExecutor
	LastError  error
	ErrorCount int
	Logger     *log.Entry
}

func (q *DnsQueryer) Exec(query *Query) Transaction {
	start := time.Now()

	results, err := q.Execute(query)
	return Transaction{
		Query:    query,
		Results:  results,
		Duration: time.Since(start),
		Error:    err,
	}
}

func (q *DnsQueryer) Validate(query *Query) error {
	return nil
}

func (q *DnsQueryer) Selectable() []string {
	return []string{"a", "cname", "mx", "ns", "ptr", "soa", "srv", "txt", "whois"}
}

func (q *DnsQueryer) Prepare(query *Query) {
	if len(query.Select) == 0 {
		return
	}

	branch := query.Select[0]
	switch branch {
	case "whois":
		query.Parts["whois"] = dns.CheckWhois(query.Target.Hostname)
	}

}

func (q *DnsQueryer) Execute(query *Query) ([]QueryResult, error) {
	var results []QueryResult
	var err error
	if len(query.Select) == 0 {
		results = []QueryResult{}
		err = ErrorNoAttrSelected
	} else {
		branch := query.Select[0]
		executor, exists := q.Executors[branch]

		if !exists {
			executor = defaultQueryDns
		}

		results, err = executor(query)
		q.LastError = err
	}

	if err != nil {
		q.ErrorCount += 1
		q.Logger.Error(err)
	} else {
		q.Logger.WithFields(log.Fields{"size": len(results)}).Info("Executed dns query")
	}
	return results, err
}

func DefaultDnsQueryer() *DnsQueryer {
	executors := map[string]dnsQueryExecutor{
		"whois": queryWhois,
	}
	return &DnsQueryer{
		Executors:  executors,
		LastError:  nil,
		ErrorCount: 0,
		Logger: log.WithFields(log.Fields{
			"struct": "DnsQueryer",
			"file":   "query/dns.go",
		}),
	}
}

type dnsQueryExecutor func(q *Query) ([]QueryResult, error)

func queryWhois(q *Query) ([]QueryResult, error) {
	details := dns.CheckWhois(q.Target.Hostname)
	result := &WhoisResult{details}

	if len(q.Select) > 1 {
		attr := q.Select[1]
		if result.HasAttr(attr) {
			return []QueryResult{&TextResult{attr, result.Attr(attr)}}, nil
		}

		return []QueryResult{}, ErrorAttrNotFound
	}

	return []QueryResult{result}, nil
}

func defaultQueryDns(q *Query) ([]QueryResult, error) {
	// @todo default NS lookup
	qtype := strings.ToUpper(q.Select[0])
	details, err := dns.CheckDNS(q.Target.AsUrl())
	results := []QueryResult{}

	if err != nil {
		panic(err)
	}
	for _, d := range details {
		if d.QueryType() == qtype {
			for _, answer := range d.Answers() {
				results = append(results, &MapResult{map[string]string{
					"type":   answer[1],
					"answer": answer[0],
				}})
			}
		}
	}
	return results, nil
}
