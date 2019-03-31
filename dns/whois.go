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
    Status        []string  `json:"status"`
    Error         error     `json:"error"`
    Raw           string    `json:"raw"`
}

func (d *Whois) ToMap() map[string]string {
    return map[string]string{
        "admin":          d.Admin,
        "admin_email":    d.AdminEmail,
        "organization":   d.Organization,
        "registrar_name": d.RegistrarName,
    }
}

func parseRecord(whois string) {
    lines := strings.Split(whois, "\n")
    for _, line := range lines {
        parts := strings.Split(line, ":")
        key := strings.TrimSpace(parts[0])
        fmt.Println(key)
    }
}

// @todo include duration of request
func CheckWhoisNet(target *url.URL) *Whois {
    hostname := target.Hostname()

    if len(hostname) == 0 && target.Scheme == "" {
        hostname = target.String()
    }

    _, hostname = IsSubDomain(hostname)

    if strings.Contains(hostname, ":") {
        hostname, _, _ = net.SplitHostPort(hostname)
    }

    record, err := whois.Fetch(hostname)

    if err != nil {
        return &Whois{Error: err}
    }

    result, _ := parser.Parse(record.String())

    createdAt, err := time.Parse(time.RFC3339, result.Registrar.CreatedDate)
    expiresAt, err := time.Parse(time.RFC3339, result.Registrar.ExpirationDate)

    details := &Whois{
        Raw:           record.String(),
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
