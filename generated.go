package query

import (
	"github.com/rezen/query/fetch"
	"github.com/rezen/query/requests/dns"
	"strconv"
	"strings"
)

type WhoisResult struct {
	Src *dns.Whois
}

func (d *WhoisResult) Attr(attr string) string {
	switch attr {
	case "admin":
		return d.Src.Admin
	case "admin_email":
		return d.Src.AdminEmail
	case "organization":
		return d.Src.Organization
	case "created_at":
		return d.Src.CreatedAt.String()
	case "expires_at":
		return d.Src.ExpiresAt.String()
	case "registrar_name":
		return d.Src.RegistrarName
	case "ns":
		return strings.Join(d.Src.NameServers, ",")
	case "dnssec":
		return d.Src.DomainDNSSEC
	case "text":
		return d.Src.Text
	case "status":
		return strings.Join(d.Src.Status, ",")
	default:
		return ""
	}
}

func (d *WhoisResult) HasAttr(attr string) bool {
	switch attr {
	case "admin":
		return true
	case "admin_email":
		return true
	case "organization":
		return true
	case "created_at":
		return true
	case "expires_at":
		return true
	case "registrar_name":
		return true
	case "ns":
		return true
	case "dnssec":
		return true
	case "text":
		return true
	case "status":
		return true
	default:
		return false
	}
}

func (d *WhoisResult) Attrs() []string {
	return []string{"admin", "admin_email", "organization", "created_at", "expires_at", "registrar_name", "ns", "dnssec", "text", "status"}
}

func (d *WhoisResult) AsText() string {
	iface := interface{}(d.Src)
	if s1, ok := iface.(interface{ AsText() string }); ok {
		return s1.AsText()
	}
	text := ""
	for _, attr := range d.Attrs() {
		text += " - " + attr + ": " + d.Attr(attr) + "\n"
	}
	return text
}

type CertificateResult struct {
	Src *fetch.Certificate
}

func (d *CertificateResult) Attr(attr string) string {
	switch attr {
	case "common_name":
		return d.Src.CommonName
	case "issuer":
		return d.Src.Issuer
	case "expiration":
		return d.Src.Expiration.String()
	case "body":
		return d.Src.Body
	case "domains":
		return strings.Join(d.Src.Domains, ",")
	case "is_wildcard":
		return strconv.FormatBool(d.Src.IsWildcard)
	case "has_expired":
		return strconv.FormatBool(d.Src.HasExpired)
	default:
		return ""
	}
}

func (d *CertificateResult) HasAttr(attr string) bool {
	switch attr {
	case "common_name":
		return true
	case "issuer":
		return true
	case "expiration":
		return true
	case "body":
		return true
	case "domains":
		return true
	case "is_wildcard":
		return true
	case "has_expired":
		return true
	default:
		return false
	}
}

func (d *CertificateResult) Attrs() []string {
	return []string{"common_name", "issuer", "expiration", "body", "domains", "is_wildcard", "has_expired"}
}

func (d *CertificateResult) AsText() string {
	iface := interface{}(d.Src)
	if s1, ok := iface.(interface{ AsText() string }); ok {
		return s1.AsText()
	}
	text := ""
	for _, attr := range d.Attrs() {
		text += " - " + attr + ": " + d.Attr(attr) + "\n"
	}
	return text
}
