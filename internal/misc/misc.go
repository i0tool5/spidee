package misc

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// FileWorker is a function to handle output file
func FileWorker(fileName string, names chan []string) {
	fmt.Println("[FW] <--Starting-->")
	out := ""
	for val := range names {
		for _, u := range val {
			out += u + "\n"
		}
	}
	err := os.WriteFile(fileName, []byte(out), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// EndsWith is a simple wrapper around strings.HasSuffix
func EndsWith(s, ending string) bool {
	return strings.HasSuffix(s, ending)
}

// GenAddr generates url based on base url and reference from it
func GenAddr(base, href string) (addr string, err error) {
	addr = "/"
	if strings.HasPrefix(href, "http") {
		addr = href
	} else {
		a, err := url.Parse(base)
		if err != nil {
			return addr, err
		}
		b, err := url.Parse(href)
		if err != nil {
			return addr, err
		}
		resolved := a.ResolveReference(b)
		addr = resolved.String()
	}
	return
}
