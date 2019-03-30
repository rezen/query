package query

import (
	"net"
	"net/url"
	"strings"
)

// There should be a concept of a LIVE target
// vs a "STATIC" target which is just a local file or
// a file in s3
type Target struct {
	Url      string
	Parsed   *url.URL
	Hostname string
	Config   map[string]string
	HttpUser string
	HttpPass string
}

// @todo get root domain for subdomains
// @todo IsWww for websites
// @todo slug
// @todo have query config info for target
// 		eg. target username/password or aws creds

func (t Target) AsUrl() *url.URL {
	return t.Parsed
}

func (t Target) RootUrl() string {
	return t.Url
}

func (t Target) IsValid() bool {
	return len(t.Url) > 0 && strings.Contains(t.Url, "http")
}

func TargetFromString(target string) Target {
	parsed, err := url.Parse(target)

	if err != nil {
		panic(err)
	}

	hostname := parsed.Host

	if strings.Contains(hostname, ":") {
		hostname, _, _ = net.SplitHostPort(hostname)
	}

	if len(hostname) == 0 {
		hostname = parsed.String()
	}

	return Target{
		Url:      target,
		Parsed:   parsed,
		Hostname: hostname,
	}
}
