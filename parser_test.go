package query

import (
	"reflect"
	"testing"
)

// @todo http > doc > img > src
// @todo dns > ssl > common-name
// @todo http > regex mailto:[^"]+
// @todo http > regex mailto:[^"]+
func TestParserBasic(t *testing.T) {
	q1 := StringToQuery("ssl")
	if q1.Resource != "ssl" {
		t.Errorf("Resource is not correct %s", q1.Resource)
	}

}

func TestParserStringToQuery(t *testing.T) {
	q1 := StringToQuery("  dns > whois > admin")
	if q1.Resource != "dns" {
		t.Errorf("Resource is not correct")
	}

	if q1.Select == nil {
		t.Errorf("Select should not be inl")
		return
	}

	if !reflect.DeepEqual(q1.Select, []string{"whois", "admin"}) {
		t.Errorf("Select is not correct")
	}
}

// https://ahermosilla.com/{{about,comments,hello}}/
// https://ahermosilla.com/{{about,comments,hello}}/

func TestParserStringToQueryWithTarget(t *testing.T) {
	// https://ahermosilla.com/{about,comments,hello}/
	q1 := StringToQuery("https://ahermosilla.com/ ~   (http .git/HEAD   > header > status_code) ~ {\"user-agent\": \"tricky-hobbit\"}")

	if !reflect.DeepEqual(q1.Select, []string{"header", "status_code"}) {
		t.Errorf("Select is not correct")
	}

	if q1.Target.Url != "https://ahermosilla.com/" {
		t.Errorf("Query target should be correct " + q1.Target.Url)
		return
	}

	if q1.Scope != ".git/HEAD" {
		t.Errorf("Query should include scope")
		return
	}

	if !reflect.DeepEqual(q1.Config, map[string]string{"user-agent": "tricky-hobbit"}) {
		t.Errorf("Did not parse the config correctly")
	}

}

func TestParserStringToQuery2(t *testing.T) {
	q1 := StringToQuery("http .git/HEAD >   header >   status_code")
	if q1.Resource != "http" {
		t.Errorf("Resource is not correct")
	}

	if q1.Scope != ".git/HEAD" {
		t.Errorf("Query should include scope")
		return
	}

	if q1.Select == nil {
		t.Errorf("Select should not be inl")
		return
	}

	if !reflect.DeepEqual(q1.Select, []string{"header", "status_code"}) {
		t.Errorf("Select is not correct")
	}
}
