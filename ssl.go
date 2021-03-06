package query

import (
	"github.com/rezen/query/ssl"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type SslQueryer struct {
	Version     string
	Name        string
	Description string
	Executors   map[string]queryExecutor
	LastError   error
	ErrorCount  int
	Logger      *log.Entry
}

func DefaultSslQueryer() *SslQueryer {
	return &SslQueryer{
		Version: "1.0.0",
		Name:    "ssl",
	}
}

func (s *SslQueryer) Prepare(q *Query) {}

func (s *SslQueryer) Validate(q *Query) error {
	if len(q.Select) == 0 {
		return ErrorMissingSelector
	}
	selector := q.Select[0]

	if selector == "*" {
		return nil
	}

	for _, v := range s.Selectable() {
		if v == selector {
			return nil
		}
	}
	return ErrorInvalidSelector
}

func (s *SslQueryer) Exec(q *Query) Transaction {
	start := time.Now()
	txn := Transaction{
		Query:    q,
		Results:  []QueryResult{},
		Duration: 0,
	}

	details, err := ssl.CheckSSL(q.Target.AsUrl())
	txn.Duration = time.Since(start)

	if err != nil {
		txn.Error = err
		return txn
	}

	selector := q.Select[0]
	if strings.Contains(selector, "chain") {
		// @todo
	}
	result := &CertificateResult{details}

	if selector == "*" {
		txn.Results = append(txn.Results, result)
	} else if result.HasAttr(selector) {
		if details.Error != nil {
			txn.Results = append(txn.Results, &TextResult{"error", details.Error.Error()})
		} else {
			txn.Results = append(txn.Results, &TextResult{selector, result.Attr(selector)})
		}
	}

	return txn
}

func (s *SslQueryer) Selectable() []string {
	selectors := (&CertificateResult{}).Attrs()
	selectors = append(selectors, "*")
	return selectors
}
