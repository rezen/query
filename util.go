package query

import (
	"math"
	"reflect"
	"strconv"
	"strings"
)

// https://www.reddit.com/r/dailyprogrammer/comments/4fc896/20160418_challenge_263_easy_calculating_shannon/
func shannonEntropy(input string) float64 {
	charMap := make(map[rune]int)
	for _, c := range input {
		if _, ok := charMap[c]; !ok {
			charMap[c] = 1
		} else {
			charMap[c]++
		}
	}

	var sum float64
	length := float64(len(input))
	for _, cnt := range charMap {
		tmp := float64(cnt) / length
		sum += tmp * math.Log2(tmp)
	}
	return -1 * sum
}

func trim(str string) string {
	str = strings.Replace(str, "\n", "", -1)
	fields := strings.Fields(str)
	str = strings.Join(fields, " ")
	return strings.TrimSpace(str)
}

func HasAttr(s interface{}, key string) bool {
	if s1, ok := s.(interface{ HasAttr(key string) bool }); ok {
		return s1.HasAttr(key)
	}

	attrs := GetAttrs(s)

	for _, attr := range attrs {
		if attr == key {
			return true
		}
	}

	return false
}

func Attr(s interface{}, key string) string {
	if s1, ok := s.(interface{ Attr(key string) string }); ok {
		return s1.Attr(key)
	}
	// f := reflect.Indirect(reflect.ValueOf(s))
	attrs := GetAttrs(s)

	for _, attr := range attrs {
		if attr == key {
			return ""
		}
	}

	return ""
}

func AsText(s interface{}) string {
	if s1, ok := s.(interface{ AsText() string }); ok {
		return s1.AsText()
	}

	// @todo if struct ... vs
	text := "x: "
	// @todo make more effecient or generate code
	t := reflect.TypeOf(s)

	f := reflect.Indirect(reflect.ValueOf(s))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := f.FieldByName(field.Name)
		value, ok := field.Tag.Lookup("json")
		fieldText := ""

		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldText = strconv.FormatInt(fv.Int(), 10)
		case reflect.String:
			fieldText = fv.String()
			// etc...
		default:
			fieldText = fv.String()
			// fmt.Println(fv.Kind())
		}

		key := ""
		if ok {
			attr := strings.Split(value, ",")[0]
			if len(attr) > 0 {
				key = attr
			}

			text += " - " + key + ": " + fieldText + "\n"
		}
	}

	return text

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
