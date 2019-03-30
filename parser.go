package query

import (
	"encoding/json"
	"regexp"
	"strings"
)

/*


// should be graphql query?
{
	http(endpoint: '/', user-agent: "chrome", data: "cats&dog", query: '&cats' ) {
		doc {
			select(img) {
				src
			}
		}
	}
}




http --user-agent={{ .VarChrome}} > doc img > first

http > doc > img > first

http > doc > img > first

// Think like linux pipes or angular 1.x filters
http | q:doc | q:img | has-alt | attr:src

// Get the domain reputation of each src
http | q:doc | q:img | to-q(src reputation.domain)


q('ssl').q('expiration')

q(ssl) | q(header) | q(status_code) | has-email

http()
	.doc('script[src]')
	.attr('src')
	.map(href => http(href).header().attr('status_code'))


query('http(.git/HEAD) > status_code').config()

q('http', '/', {"user-agent": "tricky-hobbit"})
	.select('header')
	.select('status_code')

q('ssl')
	.select('expiration').map()

ssl().expiration()
dns().whois().registrar_name()
http(".git/HEAD", {"user-agent": "tricky-hobbit"}).header().attr('status_code')

// Example with subqueres
http()
	.doc('script[src]')
	.attr('src')
	.map(function(href) {
		return http(href).header().attr('status_code')
	})


# Current setup
https://ahermosilla.com ~ (http .git/HEAD) ~ {"user-agent": "tricky-hobbit"}


var target = "https://ahermosilla.com";
var config = {
	"user-agent": "tricky-hobbit"
};
query(config, "http .git/HEAD > header > status_code", config);
query(config, "http .git/config > header > status_code", config);

// Represent query as string
// @todo need a way to include config and config
*/
func parseQuery(raw string, query *Query) *Query {
	if !strings.Contains(raw, ">") {
		query.Resource = raw
		return query
	}
	re := regexp.MustCompile(`\s+>\s+`)
	matches := re.FindAllStringSubmatchIndex(raw, -1)
	parts := []string{}
	lastIndex := 0

	for _, match := range matches {
		idx := match[0]
		parts = append(parts, raw[lastIndex:idx])
		lastIndex = match[1]
	}

	if lastIndex != 0 {
		parts = append(parts, raw[lastIndex:])
	}

	if len(parts) > 0 {
		fields := strings.Fields(parts[0])
		query.Resource = fields[0]

		if len(fields) > 1 {
			query.Scope = fields[1]
		}

		for _, part := range parts[1:] {
			query.Select = append(query.Select, strings.Fields(part)...)
		}
	}

	return query
}

func parseConfig(raw string, query *Query) *Query {
	// @todo make sure config starts with {}
	config := map[string]string{}
	err := json.Unmarshal([]byte(raw), &config)

	if err != nil {
		panic(err)
	} else {
		query.Config = config
	}

	return query
}

func StringToQuery(raw string) *Query {
	// @todo should have parsing errors
	raw = strings.TrimSpace(raw)
	query := &Query{RawSource: raw}
	if strings.Contains(raw, " ~ ") {
		// @todo needs more tests & improvements
		// Contains target in query
		parts := strings.Split(raw, " ~ ")

		for idx, part := range parts {
			parts[idx] = strings.TrimSpace(part)
		}

		qry := parts[1]
		qry = qry[1 : len(qry)-1]
		query.Target = TargetFromString(parts[0])

		if len(parts) > 2 && len(parts[2]) > 0 {
			query = parseConfig(parts[2], query)
		}
		query = parseQuery(qry, query)
	} else {
		query = parseQuery(raw, query)
	}

	return query
}
