package dns

import (
    "fmt"
    "github.com/domainr/whois"
    parser "github.com/likexian/whois-parser-go"
    "net"
    "net/url"
    "strings"
    "time"
)

// @todo JSON annotate
type Whois struct {
    Admin         string    `json:"admin"`
    AdminEmail    string    `json:"admin_email"`
    Organization  string    `json:"organization"`
    CreatedAt     time.Time `json:"created_at"` // @todo change to date types
    ExpiresAt     time.Time `json:"expires_at"`
    RegistrarName string    `json:"registrar_name"`
    NameServers   []string  `json:"ns"`
    DomainDNSSEC  string    `json:"dnssec"`
    Text          string    `json:"text"`
    Status        []string  `json:"status"`
}

func (d *Whois) AsText() string {
    return d.Text
}

func (d *Whois) ToMap() map[string]string {
    return map[string]string{
        "admin":          d.Admin,
        "admin_email":    d.AdminEmail,
        "organization":   d.Organization,
        "registrar_name": d.RegistrarName,
    }
}

// @todo include duration of request
func CheckWhoisNet(target *url.URL) *Whois {
    _, hostname := IsSubDomain(target.Host)

    if strings.Contains(hostname, ":") {
        hostname, _, _ = net.SplitHostPort(hostname)
    }

    record, _ := whois.Fetch(hostname)
    result, _ := parser.Parse(record.String())

    // fmt.Println(record.String())
    fmt.Println("")

    createdAt, err := time.Parse(time.RFC3339, result.Registrar.CreatedDate)
    expiresAt, err := time.Parse(time.RFC3339, result.Registrar.ExpirationDate)

    fmt.Println(err)

    details := &Whois{
        Text:          record.String(),
        Admin:         result.Admin.Name,
        AdminEmail:    result.Admin.Email,
        Organization:  result.Admin.Organization,
        RegistrarName: result.Registrar.RegistrarName,
        CreatedAt:     createdAt,
        ExpiresAt:     expiresAt,
        NameServers:   strings.Split(result.Registrar.NameServers, ","),
        Status:        strings.Split(result.Registrar.DomainStatus, ","),
        DomainDNSSEC:  result.Registrar.DomainDNSSEC,
    }

    return details
}

func CheckWhois(domain string) *Whois {
    parsed, _ := url.Parse(domain)
    return CheckWhoisNet(parsed)
}
