package generators

import (
	"bytes"
	"fmt"
	"github.com/rezen/query/dns"
	"github.com/rezen/query/ssl"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"
	"time"
)

// @todo for every type of result generate
// interface fulfillment if it doesn't have it already
// https://www.calhoun.io/using-code-generation-to-survive-without-generics-in-go/
// https://blog.carlmjohnson.net/post/2016-11-27-how-to-use-go-generate/

var header = `
package query

import (
	"github.com/rezen/query/ssl"
	"github.com/rezen/query/dns"
	"strings"
	"strconv"
)`

var templateAttr = `

type {{ .Name }}Result struct {
	Src *{{ .FullName }}
}

func (d *{{ .Name }}Result) Attr(attr string) string {
	switch attr { {{range .Fields}}
	case "{{ .Accessor }}":
		return {{ .Name }}{{end}}
	default:
		return ""
	}
}

func (d *{{ .Name }}Result) HasAttr(attr string) bool {
	switch attr { {{range .Fields}}
	case "{{ .Accessor }}":
		return true{{end}}
	default:
		return false
	}
}

func (d *{{ .Name }}Result) Attrs() []string {
	return []string{ {{ range .Attrs}}"{{.}}",{{ end }}}
}

func (d *{{ .Name }}Result) AsText() string {
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

`

type Field struct {
	Accessor string
	Name     string
}

type GenerateResults struct {
	Name     string // type of result
	FullName string
	Fields   []Field
	Attrs    []string
}

func GetAttrs(s interface{}) []string {
	if s1, ok := s.(interface{ Attrs() []string }); ok {
		return s1.Attrs()
	}
	t := reflect.TypeOf(s)
	attrs := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		value, ok := field.Tag.Lookup("json")

		if ok {
			attr := strings.Split(value, ",")[0]
			if len(attr) > 0 {
				attrs = append(attrs, attr)
			}
		}
	}

	return attrs
}

func GenerateStuff() {
	types := []interface{}{
		dns.Whois{},
		ssl.Certificate{},
	}

	tpl := bytes.NewBuffer([]byte(header))

	for _, model := range types {
		fmt.Println()
		if d, ok := model.(interface{ Attr(string) string }); ok {
			fmt.Println(d)
		} else {
			t := reflect.TypeOf(model)
			fmt.Println(reflect.TypeOf(GenerateResults{}).PkgPath())
			z := GenerateResults{
				Name:     t.Name(),
				FullName: t.String(),
				Attrs:    GetAttrs(model),
			}
			fields := []Field{}

			// @todo sort fields
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				value, ok := field.Tag.Lookup("json")

				if ok {
					attr := strings.Split(value, ",")[0]
					if len(attr) > 0 {
						getter := "d.Src." + field.Name

						if field.Type.Name() == "error" {
							getter += ".Error()"
						}
						switch field.Type {
						case reflect.TypeOf(time.Time{}):
							getter = getter + ".String()"
						case reflect.TypeOf([]string{}):
							getter = "strings.Join(" + getter + ", \",\")"
						case reflect.TypeOf(true):
							getter = "strconv.FormatBool(" + getter + ")"
						default:

							_, ok := field.Type.MethodByName("String")
							if ok {
								fmt.Println("yay", field.Type)
								getter += ".String()"
							}
						}

						fields = append(fields, Field{
							Name:     getter,
							Accessor: attr,
						})
					}
				}
			}

			z.Fields = fields

			tmpl, err := template.New("test").Parse(templateAttr)

			if err != nil {
				panic(err)
			}
			if err := tmpl.Execute(tpl, z); err != nil {
				fmt.Println(err)
			}

			ioutil.WriteFile("../generated.go", tpl.Bytes(), 0666)

		}
	}

	// fmt.Println(tpl.String())

}
