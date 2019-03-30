package requests

import (
	// "bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func sortAndHash(list []string) string {
	sort.Strings(list)
	hash := strings.Join(list, " ")
	hasher := md5.New() // change to sha1?
	hasher.Write([]byte(hash))
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)

}

func isIgnoreableResHeader(header string) bool {
	header = strings.ToLower(header)
	ignoreable := []string{
		"connection",
		"date",
		"vary",
		"expires",
		"last-modified",
		"cache-control",
		"pragma",
		"etag",
	}

	for _, val := range ignoreable {
		if val == header {
			return true
		}
	}

	return false
}

type Transaction struct {
	Request         *http.Request  `json:"-"`
	Response        *http.Response `json:"-"`
	ResponseBody    string         `json:"-"`
	ResponseHash    string         `json:"-"`
	Errors          []error        `json:"-"`
	Target          *url.URL       `json:"-"`
	Redirects       int            `json:"redirects"`
	Duration        time.Duration  `json:"-"`
	RedirectsHttp   bool           `json:"-"`
	FirstStatusCode int            `json:"first_status_code"`
}

type TransactionJson struct {
	Transaction
	Url         string            `json:"url"`
	EndUrl      string            `json:"end_url"`
	Id          string            `json:"id"`
	Headers     map[string]string `json:"headers"`
	Cookies     []string          `json:"cookies"`
	BodyHash    string            `json:"body_hash"`
	StatusCode  int               `json:"status_code"`
	HeadersHash string            `json:"headers_hash"`
}

func (t Transaction) ToJSON() ([]byte, error) {
	/*
		buffer := new(bytes.Buffer)
		t.Request.Write(buffer)
		fmt.Println(buffer)
	*/

	headers := t.Headers()
	targetUrl := t.Url().String()
	endUrl := t.Response.Request.URL.String()

	if targetUrl == endUrl {
		endUrl = ""
	}

	for header := range headers {
		if isIgnoreableResHeader(header) {
			delete(headers, header)
		}
	}

	tj := TransactionJson{
		Transaction: t,
		Url:         targetUrl,
		EndUrl:      endUrl,
		Id:          t.Id(),
		HeadersHash: t.HeadersHash(),
		Headers:     headers,
		BodyHash:    t.BodyHash(),
		Cookies:     []string{},
		StatusCode:  t.StatusCode(),
	}

	return json.Marshal(tj)
}

func TransactionFromUrl(path string) Transaction {
	target, _ := url.Parse(path)
	return Transaction{
		Target:          target,
		FirstStatusCode: 0,
	}
}

func (t Transaction) Body() string {
	return t.ResponseBody
}

func (t Transaction) Url() *url.URL {
	return t.Target
}

func (t Transaction) HeaderExists(key string) bool {
	headers := t.Headers()
	_, exist := headers[key]
	return exist
}

// Response headers
func (t Transaction) Header(key string) string {
	header := t.Response.Header.Get(strings.ToLower(key))
	return header
}

// Response headers
func (t Transaction) HeadersHash() string {
	parts := []string{}
	for key, value := range t.Headers() {

		if isIgnoreableResHeader(key) {
			continue
		}

		parts = append(parts, key)
		parts = append(parts, value)
	}
	return sortAndHash(parts)
}

func (t Transaction) BodyHash() string {
	return t.ResponseHash
}

func (t Transaction) Slug() string {
	return t.Request.Method + "--" + t.Url().String() + "--" + t.RequestHash()
}

func (t Transaction) Id() string {
	return t.RequestHash()
}

func (t Transaction) Headers() map[string]string {
	headers := map[string]string{}
	if t.Response == nil {
		return headers
	}
	// @todo fix for multiple set-cookie
	for key, value := range t.Response.Header {
		headers[strings.ToLower(key)] = value[0]
	}
	return headers
}

func (t Transaction) RequestHash() string {
	parts := []string{}
	parts = append(parts, t.Request.Method)
	parts = append(parts, t.Url().String())

	for name, values := range t.Request.Header {
		if name == "Cache-Control" {
			continue
		}

		if name == "Referer" {
			continue
		}

		if name == "Upgrade-Insecure-Requests" {
			continue
		}

		// @todo what happens when server sets cookie?
		if name == "Cookie" {
			continue
		}

		parts = append(parts, name)
		for _, value := range values {
			parts = append(parts, value)
		}
	}

	return sortAndHash(parts)
}

func (t Transaction) StatusCode() int {
	if len(t.Errors) > 0 {
		return 0
	}
	return t.Response.StatusCode
}

func (t Transaction) IsScript() bool {
	return strings.Contains(t.Header("content-type"), "javascript")
}

func (t Transaction) IsStylesheet() bool {
	return strings.Contains(t.Header("content-type"), "css")
}

func (t Transaction) IsHtml() bool {
	return strings.Contains(t.Header("content-type"), "html")
}

func (t Transaction) IsImage() bool {
	return strings.Contains(t.Header("content-type"), "image/")
}

func (t Transaction) WasRedirected() bool {
	return t.Redirects > 0
}
