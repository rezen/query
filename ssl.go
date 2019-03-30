package query

import (
	"fmt"
	"github.com/rezen/query/fetch"
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
	return nil
	// return ErrorInvalidSelector
}

func (s *SslQueryer) Exec(q *Query) Transaction {
	start := time.Now()
	txn := Transaction{
		Query:    q,
		Results:  []QueryResult{},
		Duration: 0,
	}

	details, err := fetch.CheckSSL(q.Target.AsUrl())
	txn.Duration = time.Since(start)

	if err != nil {
		fmt.Println(err)
		txn.Error = err
		return txn
	}

	selector := q.Select[0]
	if strings.Contains(selector, "chain") {
		// @todo
	}

	switch selector {
	case "issuer":
		txn.Results = append(txn.Results, &TextResult{selector, details.Issuer})
	case "expiration":
		txn.Results = append(txn.Results, &TextResult{selector, details.Expiration.String()})
	case "common-name":
		txn.Results = append(txn.Results, &TextResult{selector, details.CommonName})
	case "body":
		txn.Results = append(txn.Results, &TextResult{selector, details.Body})
	case "*":
		txn.Results = append(txn.Results, &CertificateResult{&details})
	default:
		txn.Results = append(txn.Results, &CertificateResult{&details})
	}

	return txn
}

func (s *SslQueryer) Selectable() []string {
	return []string{"issuer", "expiration", "common-name", "*"}
}
