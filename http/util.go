package http

import (
	"fmt"
)

type infermap map[string]interface{}

func (m infermap) String(key string) string {
	value := ""
	if m[key] != nil {
		value = m[key].(string)
	}

	return value
}

func (m infermap) Float(key string) float64 {
	value := 0.0
	if m[key] != nil {
		value = m[key].(float64)
	}

	return value
}

func (m infermap) MapStringString(key string) map[string]string {
	value := map[string]string{}
	if m[key] != nil {
		tmp := m[key].(map[string]interface{})
		for k, v := range tmp {
			value[k] = fmt.Sprintf("%v", v) // @todo improve
		}
	}

	return value
}
