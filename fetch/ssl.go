package fetch

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"net/url"
	"strings"
	"time"
)

type Certificate struct {
	CommonName string    `json:"common_name"`
	Issuer     string    `json:"issuer"`
	Expiration time.Time `json:"expiration"`
	Body       string    `json:"body"`
	Domains    []string  `json:"domains"`
	IsWildcard bool      `json:"is_wildcard"`
	HasExpired bool      `json:"has_expired"`
}

func DetailsFromChain(chain []*x509.Certificate) Certificate {
	cert := chain[0]
	commonName := cert.Subject.CommonName
	var data bytes.Buffer
	pem.Encode(&data, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

	return Certificate{
		Body:       data.String(),
		CommonName: commonName,
		Issuer:     strings.Join(cert.Issuer.Organization, " "),
		Domains:    cert.DNSNames,
		Expiration: cert.NotAfter,
		HasExpired: time.Now().After(cert.NotAfter),
		IsWildcard: strings.Contains(commonName, "*"),
	}
}

// https://github.com/mozilla/tls-observatory
func CheckSSL(target *url.URL) (Certificate, error) {
	host := target.Host
	if !strings.Contains(host, ":") {
		host = host + ":443"
	}

	dialer := &net.Dialer{Timeout: 3 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", host, nil)
	defer conn.Close()

	if err != nil {
		return Certificate{}, err
	}

	chain := conn.ConnectionState().VerifiedChains
	details := DetailsFromChain(chain[0])
	return details, nil
}
