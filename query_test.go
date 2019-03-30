package query

import (
	"fmt"
	"testing"
)

func TestQuery2(t *testing.T) {
	// @todo use @text for attributes?
	example := &Query{
		Target:   TargetFromString("https://ahermosilla.com"),
		Resource: "http",
		Scope:    "",
		Select:   []string{"doc", "meta"},
	}

	transaction := Search(example)

	for _, result := range transaction.Results {
		fmt.Println("--- query-attr-name", result.Attr("name"))
	}
}

func TestQueryZ(t *testing.T) {
	// @todo use @text for attributes?
	example := &Query{
		Target:   TargetFromString("https://ahermosilla.com"),
		Resource: "http",
		Scope:    "/cats",
		Select:   []string{"doc", "title"},
	}
	transaction := Search(example)

	fmt.Println("--- title", transaction.Results[0])
}

func TestQueryDocComments(t *testing.T) {
	example := &Query{
		Target:   TargetFromString("https://jpmktg.com"),
		Resource: "http",
		Select:   []string{"doc", "comments"},
	}
	txn := Search(example)

	for _, result := range txn.GetResults() {
		fmt.Println(result)
	}
	t.Fail()

}
func TestQuery3(t *testing.T) {

	example := &Query{
		Target:   TargetFromString("https://www.ahermosilla.com"),
		Resource: "dns",
		Scope:    "",
		Select:   []string{"whois", "registrar_name"},
	}
	fmt.Println(Search(example))
}

func TestQueryBad(t *testing.T) {
	queryer := DefaultQueryer()
	fmt.Println(queryer.Selectable())
	t.Fail()
}
