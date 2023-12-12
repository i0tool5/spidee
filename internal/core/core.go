package core

import (
	"io"
	"log"
	"net/http"
	"regexp"
)

// Fetcher interface
type Fetcher interface{}

// Fetched it's a structure which represents
// data retrieved from remote source
type Fetched struct {
	baseURL string
	hrefs   []string
}

// FetchedArr represents list of fetched elements
type FetchedArr []Fetched

// Hrefs returns list of href elements
func (f *Fetched) Hrefs() []string {
	return f.hrefs
}

// Base returns origin url
func (f *Fetched) Base() string {
	return f.baseURL
}

func Fetch(url string) (*Fetched, error) {
	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	dat, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	p := searchLinks(dat)

	return &Fetched{baseURL: url, hrefs: p}, nil
}

// searchLinks searches for links on the page and returns them
func searchLinks(data []byte) []string {
	hrefs := make([]string, 0)
	hrefRegex, err := regexp.Compile(
		`<a\s+.*href=["|']{0,1}(https?://[A-Za-z0-9\.\/\?=\-_]+|` +
			`/?[A-Za-z0-9_/]+/?)["|']{0,1}\s?.*>`,
	)
	if err != nil {
		log.Fatal(err)
	}

	srcRegex, err := regexp.Compile(
		`<img\s+.*src=["|']{0,1}([A-Za-z0-9\.\/\?=\-_:]+)["|']{0,1}\s?.*>`,
	)
	if err != nil {
		log.Fatal(err)
	}
	found := hrefRegex.FindAllSubmatch(data, -1)
	for _, v := range found {
		u := string(v[len(v)-1])
		hrefs = append(hrefs, u)
	}

	found = srcRegex.FindAllSubmatch(data, -1)
	for _, v := range found {
		u := string(v[len(v)-1])
		hrefs = append(hrefs, u)
	}
	return hrefs
}
