package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

/*
* Fetched it's a structure which represents
*  data retrieved from remote source
 */
type Fetched struct {
	baseURL string
	hrefs   []string
}

type FetchedArr []Fetched

func (f *Fetched) GetHrefs() []string {
	return f.hrefs
}

func (f *Fetched) GetBase() string {
	return f.baseURL
}

func Fetch(url string) Fetched {
	req, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	dat, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	req.Body.Close()
	p := parsePage(dat)

	return Fetched{baseURL: url, hrefs: p}
}

func parsePage(data []byte) []string {
	r := make([]string, 0)
	href, err := regexp.Compile(`<a\s+.*href=["|']{0,1}(https?://[A-Za-z0-9\.\/\?=\-_]+|/?[A-Za-z0-9_/]+/?)["|']{0,1}\s?.*>`)
	if err != nil {
		log.Fatal(err)
	}

	src, err := regexp.Compile(`<img\s+.*src=["|']{0,1}([A-Za-z0-9\.\/\?=\-_:]+)["|']{0,1}\s?.*>`)
	if err != nil {
		log.Fatal(err)
	}
	found := href.FindAllSubmatch(data, -1)
	for _, v := range found {
		u := fmt.Sprintf("%s", v[len(v)-1])
		r = append(r, u)
	}

	found = src.FindAllSubmatch(data, -1)
	for _, v := range found {
		u := fmt.Sprintf("%s", v[len(v)-1])
		r = append(r, u)
	}
	return r
}
