package dns

import (
    // lru "github.com/hashicorp/golang-lru"
    "github.com/miekg/dns"
    "golang.org/x/net/publicsuffix"

    "fmt"
    "net"
    "net/url"
    "strings"
    "time"
)

type DnsQuery struct {
    query   *dns.Msg
    answers [][]string
}

type Answer struct {
    RecordType string
}

func (d *DnsQuery) QueryType() string {
    // github.com/miekg/dns/types.go
    switch d.query.Question[0].Qtype {
    case dns.TypeNS:
        return "NS"
    case dns.TypeA:
        return "A"
    case dns.TypeCNAME:
        return "CNAME"
    case dns.TypeTXT:
        return "TXT"
    case dns.TypeSOA:
        return "SOA"
    case dns.TypeSRV:
        return "SRV"
    case dns.TypeMX:
        return "MX"

    }

    return "????"
}

func (d *DnsQuery) Answers() [][]string {
    return d.answers
}

func (d *DnsQuery) AsMap() map[string]string {
    answer := ""
    for _, parts := range d.answers {
        answer += strings.Join(parts, ",")
    }
    return map[string]string{
        "qtype":   d.QueryType(),
        "domain":  d.query.Question[0].Name,
        "answers": answer,
    }
}

func (d *DnsQuery) Id() string {
    return d.QueryType() + ":" + d.query.Question[0].Name
}

func DetailsFromQuestion(q *dns.Msg) *DnsQuery {
    return &DnsQuery{q, [][]string{}}
}

func (d *DnsQuery) AddAnswer(answer interface{}) {
    ans := []string{"?", "?"}

    switch v := answer.(type) {
    case *dns.TXT:
        ans = []string{strings.Join(answer.(*dns.TXT).Txt, ";"), "TXT"}
    case *dns.NS:
        ans = []string{answer.(*dns.NS).Ns, "NS"}
    case *dns.MX:
        ans = []string{answer.(*dns.MX).Mx, "MX"}
    case *dns.PTR:
        ans = []string{answer.(*dns.PTR).Ptr, "PTR"}
    case *dns.A:
        ans = []string{answer.(*dns.A).A.String(), "A"}
    case *dns.CNAME:
        ans = []string{answer.(*dns.CNAME).Target, "CNAME"}
    case *dns.SRV:
        ans = []string{answer.(*dns.SRV).Target, "SRV"} // port?
    case *dns.SOA:
        record := answer.(*dns.SOA)
        ans = []string{record.Ns + ";" + record.Mbox, "SOA"}
    default:
        fmt.Printf("I don't know about type %T!\n", v)
    }

    d.answers = append(d.answers, ans)
}

type DnsRequestor struct {
    Config *DnsConfig
}

func (d *DnsRequestor) Client() dns.Client {
    return dns.Client{
        ReadTimeout: time.Duration(d.Config.Timeout) * time.Second,
    }
}

func questionNs(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeNS)
    return message
}

func questionA(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeA)
    return message
}

func questionCname(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeCNAME)
    return message
}

func questionTxt(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeTXT)
    return message
}

func questionMx(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeMX)
    return message
}

func questionSoa(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeSOA)
    return message
}

func questionSrv(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypeSRV)
    return message
}

func questionPtr(hostname string) *dns.Msg {
    message := &dns.Msg{}
    message.SetQuestion(hostname+".", dns.TypePTR)
    return message
}

type DnsConfig struct {
    Server    string
    Timeout   int
    TryDNSSEC bool
}

func DefaultDnsConfig() *DnsConfig {
    return &DnsConfig{
        Server:    "8.8.8.8:53",
        Timeout:   8,
        TryDNSSEC: false,
    }
}

func IsSubDomain(hostname string) (bool, string) {
    suffix, _ := publicsuffix.PublicSuffix(hostname)
    extra := strings.Replace(hostname, "."+suffix, "", -1)
    parts := strings.Split(extra, ".")
    return len(parts) > 1, parts[len(parts)-1] + "." + suffix
}

func CheckDNSRecord(target *url.URL, qtype string) (*DnsQuery, error) {
    config := DefaultDnsConfig()
    hostname := target.Host
    requestor := &DnsRequestor{config}
    client := requestor.Client()

    if strings.Contains(hostname, ":") {
        hostname, _, _ = net.SplitHostPort(hostname)
    }

    ip := net.ParseIP(hostname)

    // If the hostname is actually an IP, ignore it
    if ip != nil {
        return &DnsQuery{}, nil
    }

    fmt.Println("Hostname", hostname)
    var query *DnsQuery
    switch qtype {
    case "ns":
        query = DetailsFromQuestion(questionNs(hostname))
    case "a":
        query = DetailsFromQuestion(questionA(hostname))
    case "txt":
        query = DetailsFromQuestion(questionTxt(hostname))
    case "cname":
        query = DetailsFromQuestion(questionCname(hostname))
    case "mx":
        query = DetailsFromQuestion(questionMx(hostname))
    case "soa":
        query = DetailsFromQuestion(questionSoa(hostname))
    case "srv":
        query = DetailsFromQuestion(questionSrv(hostname))

    }
    response, _, err := client.Exchange(query.query, requestor.Config.Server)

    if err != nil {
        return query, err
    }
    for _, answer := range response.Answer {
        query.AddAnswer(answer)
    }
    return query, err
}

func CheckDNS(target *url.URL) ([]*DnsQuery, error) {
    config := DefaultDnsConfig()
    hostname := target.Host
    requestor := &DnsRequestor{config}
    client := requestor.Client()

    if strings.Contains(hostname, ":") {
        hostname, _, _ = net.SplitHostPort(hostname)
    }

    if len(hostname) == 0 {
        hostname = target.String()
    }

    ip := net.ParseIP(hostname)

    // If the hostname is actually an IP, ignore it
    if ip != nil {
        return []*DnsQuery{}, nil
    }

    isSub, domainRoot := IsSubDomain(hostname)

    // https://help.github.com/articles/using-a-custom-domain-with-github-pages/
    queries := []*DnsQuery{
        DetailsFromQuestion(questionNs(hostname)),
        DetailsFromQuestion(questionA(hostname)),
        DetailsFromQuestion(questionTxt(hostname)),
        DetailsFromQuestion(questionCname(hostname)),
        DetailsFromQuestion(questionMx(hostname)),
        DetailsFromQuestion(questionSoa(hostname)),
        DetailsFromQuestion(questionSrv(hostname)),
        DetailsFromQuestion(questionPtr(hostname)),
    }

    if isSub {
        queries = append(queries, DetailsFromQuestion(questionCname(domainRoot)))
        queries = append(queries, DetailsFromQuestion(questionNs(domainRoot)))
    }

    for _, qry := range queries {
        // @todo cache query
        response, _, err := client.Exchange(qry.query, requestor.Config.Server)
        if err != nil {
            continue
        }

        for _, answer := range response.Answer {
            qry.AddAnswer(answer)
        }
    }

    return queries, nil
}
