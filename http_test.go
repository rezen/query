package query

import (
	"fmt"
	"testing"
)

func TestHttpFirst(t *testing.T) {
	http := DefaultHttpQueryer()
	fmt.Println(http.Selectable())
}
