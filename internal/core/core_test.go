package core

import (
	"reflect"
	"testing"
)

func Test_parsePage(t *testing.T) {
	t.Run("success with anchor link", func(t *testing.T) {
		want := []string{"http://example.com/page1.htm"}
		data := []byte(`<a href="http://example.com/page1.htm">Link</a>`)

		got := searchLinks(data)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("parsePage() = %v, want %v", got, want)
		}
	})
}
