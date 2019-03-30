package ssl

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
	Error      error     `json:"error"`
}

func DetailsFromChain(chain []*x509.Certificate) Certificate {
	cert := chain[0]
	commonName := cert.Subject.CommonName
	var data bytes.Buffer
	pem.Encode(&data, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

	certificate := Certificate{
		Body:       data.String(),
		CommonName: commonName,
		Issuer:     strings.Join(cert.Issuer.Organization, " "),
		Domains:    cert.DNSNames,
		Expiration: cert.NotAfter,
		HasExpired: time.Now().After(cert.NotAfter),
		IsWildcard: strings.Contains(commonName, "*"),
	}

	if certificate.IsWildcard {
		return certificate
	}

	for _, domain := range certificate.Domains {
		if strings.Contains(domain, "*") {
			certificate.IsWildcard = true
			return certificate
		}
	}
	return certificate
}

// https://github.com/mozilla/tls-observatory
func CheckSSL(target *url.URL) (Certificate, error) {
	var conn *tls.Conn
	var err1 error
	var err2 error
	var details Certificate

	host := target.Host

	if !strings.Contains(host, ":") {
		host = host + ":443"
	}

	dialer := &net.Dialer{Timeout: 3 * time.Second}
	// VerifyPeerCertificate
	conn, err1 = tls.DialWithDialer(dialer, "tcp", host, nil)

	// @todo there has to be a better way than redialing
	if err1 != nil {
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err2 = tls.DialWithDialer(dialer, "tcp", host, config)
	}

	defer conn.Close()

	chain := conn.ConnectionState().VerifiedChains
	chain2 := conn.ConnectionState().PeerCertificates
	if len(chain) > 0 {
		details = DetailsFromChain(chain[0])

	} else {
		details = DetailsFromChain(chain2)
	}

	if err1 != nil {
		details.Error = err1

	} else if err2 != nil {
		details.Error = err2
	}
	return details, nil
}
