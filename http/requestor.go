package http

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"time"
)

// @todo streamline extras
// @todo build history module!
// @todo if request is currently queued up, wait for it
type Response interface {
	Url() *url.URL
	Body() string
	Header(key string) string
	Headers() map[string]string
	StatusCode() int
}

type Requestor struct {
	Target  string
	Base    *url.URL
	History map[string]*Transaction
	ByHash  map[string]*Transaction
	Config  map[string]interface{}
	Jar     *cookiejar.Jar
}

type UrlStatus struct {
	URL        *url.URL
	StatusCode int
}

type tweakedTransport struct {
	*http.Transport
	StatusCodes []UrlStatus
}

func (t *tweakedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.Transport.RoundTrip(request)
	statusCode := 0
	if err == nil {
		statusCode = response.StatusCode
	}
	t.StatusCodes = append(t.StatusCodes, UrlStatus{request.URL, statusCode})
	return response, err
}

func getTransport(config *RequestorConfig) *tweakedTransport {
	// @todo option to proxy with tor
	// https://www.devdungeon.com/content/making-tor-http-requests-go
	statusCodes := make([]UrlStatus, 0)

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if len(config.Proxy) > 3 {
		proxy, err := url.Parse(config.Proxy)
		if err != nil {
			fmt.Println("fail")
		}
		transport.Proxy = http.ProxyURL(proxy)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// @todo log
	} else if len(os.Getenv("HTTP_PROXY")) > 3 {
		// @todo way to have custom certificate
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// @todo log
	}

	if config.IgnoreBadSSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &tweakedTransport{transport, statusCodes}
}

func (r *Requestor) SetRequestHeaders(req *http.Request, config *RequestorConfig) {
	if config.UseAuthentication {
		// @todo move to user?
		auth := config.Username + ":" + config.Password
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Add("Authorization", "Basic "+encoded)
	}

	for header, value := range config.Headers {
		req.Header.Set(header, value)
	}
}

// @todo emit request event
func (r *Requestor) Get(path string) *Transaction {

	// @todo history should use hash vs path
	if _, ok := r.History[path]; ok {
		return r.History[path]
	}

	redirects := 0
	errors := make([]error, 0)
	page, _ := url.Parse(path)
	updated := r.Base.ResolveReference(page)
	config := CreateDefaultConfig()
	start := time.Now()
	timeout := time.Duration(config.Timeout) * time.Second
	request, err := http.NewRequest("GET", updated.String(), nil)

	r.SetRequestHeaders(request, config)

	// @todo bail
	if err != nil {
		errors = append(errors, err)
	}

	// @todo cache transport instead of rebuilding
	transport := getTransport(config)
	client := http.Client{
		Timeout:   timeout,
		Jar:       r.Jar,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !config.FollowRedirects {
				return http.ErrUseLastResponse
			}

			// @todo emit event
			request = req
			redirects += 1

			if redirects <= 3 {
				r.SetRequestHeaders(req, config)
				return nil
			}
			return http.ErrUseLastResponse
		},
	}
	address := ""
	// https://blog.golang.org/http-tracing
	trace := &httptrace.ClientTrace{
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {},
		GotConn: func(connInfo httptrace.GotConnInfo) {},
		ConnectDone: func(network, addr string, err error) {
			address = addr
		},
	}
	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

	response, err := client.Do(request)

	if err != nil {
		// @todo log errs
		errors = append(errors, err)
	}

	firstCode := 0

	if len(transport.StatusCodes) > 0 {
		firstCode = transport.StatusCodes[0].StatusCode
	}

	// http://www.debug.is/2017/11/01/http-tls-handling-go-http-client/
	transaction := &Transaction{
		Target:          updated,
		Request:         request,
		Response:        response,
		Errors:          errors,
		Redirects:       redirects,
		Duration:        time.Since(start),
		FirstStatusCode: firstCode,
		RedirectsHttp:   strings.Contains(request.URL.String(), "https:") && strings.Contains(updated.String(), "http:"),
		IP:              address,
	}

	if err == nil {
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		transaction.ResponseBody = string(bodyBytes)
		hasher := md5.New()
		hasher.Write(bodyBytes)
		hashed := hasher.Sum(nil)
		transaction.ResponseHash = "md5:" + hex.EncodeToString(hashed)
	}

	r.History[path] = transaction
	return transaction
}

func CreateRequestorWithTarget(target string) *Requestor {
	parsed, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	transactions := map[string]*Transaction{}
	config := map[string]interface{}{}

	// @todo use file for cookie jar
	jar, _ := cookiejar.New(nil)

	return &Requestor{
		Target:  target,
		Base:    parsed,
		History: transactions,
		Config:  config,
		Jar:     jar,
	}
}

type RequestorConfig struct {
	Headers           map[string]string
	UseAuthentication bool
	Username          string
	Password          string
	FollowRedirects   bool
	Timeout           int
	IgnoreBadSSL      bool
	Debug             bool
	Proxy             string
}

func ConfigFromMap(map[string]string) *RequestorConfig {
	config := CreateDefaultConfig()
	return config
}

func CreateDefaultConfig() *RequestorConfig {
	headers := map[string]string{
		"X-Awwwditor":               "1",
		"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.90 Safari/537.36",
		"Cache-Control":             "max-age=0",
		"Upgrade-Insecure-Requests": "1",
	}

	return &RequestorConfig{
		Headers:           headers,
		UseAuthentication: false,
		Username:          "",
		Password:          "",
		FollowRedirects:   true,
		Timeout:           4,
		IgnoreBadSSL:      false,
		Debug:             false,
	}
}
