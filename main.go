package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CLI arguments variables
var (
	dpth         int
	netRoutines  int
	parsRoutines int
	begin        string
	outFile      string
	formats      string
	ignore       string
)

type Constraints struct {
	svFmts  []string
	ignored []string
	fileOut string
}

// Crawler is a main struct for crawling web sites
type Crawler struct {
	depth    int
	mainURL  string
	visited  map[string]bool
	locker   chan bool
	startURL string
	netCoro  int
	parsCoro int
	Constraints
}

// Fetched a
type Fetched struct {
	baseURL string
	hrefs   []string
}

func fetch(url string) Fetched {
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

func fileWorker(fileName string, names chan []string) {
	fmt.Println("[FW] <--Starting-->")
	out := ""
	for val := range names {
		for _, u := range val {
			out += u + "\n"
		}
	}
	err := ioutil.WriteFile(fileName, []byte(out), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func endsWith(s, ending string) bool {
	return strings.HasSuffix(s, ending)
}

func (c *Crawler) startFetchers(ctx context.Context, inpCh chan []string, outCh chan []Fetched) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[FETCHERS] Signal Done.")
			close(outCh)
			return
		case urls, ok := <-inpCh:
			if !ok {
				close(outCh)
				return
			}
			z := make([]Fetched, 0)
			inside := make(chan string)
			dn := make(chan bool)
			go func() {
				// filling channel with urls and sends it to `fetcher`
				for _, url := range urls {
					if !c.isIgnored(url) {
						inside <- url
					}
				}
				close(inside)
			}()
			for i := 0; i < c.netCoro; i++ {
				go func(n int, c chan string, d chan bool) {
					for url := range c {
						fmt.Printf("[F %d] fetching %s\n", n, url)
						dat := fetch(url)
						z = append(z, dat)
					}
					d <- true
				}(i, inside, dn)
			}
			for i := 0; i < c.netCoro; i++ {
				<-dn
			}
			outCh <- z
		}
	}
}

func (c *Crawler) startCheckers(cancelFunc func(), ch chan []Fetched, fout, out chan []string) {
	done := make(chan bool, c.parsCoro)
	for i := 0; i < c.parsCoro; i++ {
		go func(n int, cf chan []Fetched, tf, o chan []string, d chan bool) {
			for {
				fetchedArr, ok := <-cf
				if !ok {
					d <- true
					return
				}
				urls := make([]string, 0)
				tout := make([]string, 0)
				addr := "/"
				for _, fStruct := range fetchedArr {
					hrefs := fStruct.hrefs
					for _, href := range hrefs {
						switch {
						case strings.HasPrefix(href, "http"):
							addr = href
						default:
							addr = fStruct.baseURL + href
						}
						if !c.visit(addr) {
							urls = append(urls, addr)
							fmt.Printf("[C %d]->> %s\n", n, addr)
							for _, ending := range c.svFmts {
								if endsWith(addr, ending) {
									tout = append(tout, addr)
								}
							}
						}
					}
				}
				if len(c.svFmts) > 0 && c.fileOut != "" {
					tf <- tout

				}
				if c.depth <= 0 {
					fmt.Println("[HANDLER] Depth exceeded!")
					cancelFunc()
					d <- true
					return
				}
				o <- urls
				c.depth--
			}
		}(i, ch, fout, out, done)
	}
	for i := 0; i < c.parsCoro; i++ {
		<-done
	}
	fmt.Println("[Checkers] Done!")
	close(fout)
	close(out)
}

func (c *Crawler) isIgnored(url string) bool {
	for _, ig := range c.ignored {
		if strings.Contains(url, ig) && ig != "" {
			return true
		}
	}
	return false
}

func (c *Crawler) visit(url string) bool {
	c.locker <- true
	resp := false
	if c.visited[url] {
		resp = true
	}
	c.visited[url] = true
	<-c.locker
	return resp
}

func (c *Crawler) Crawl() {
	beginTime := time.Now()
	u := c.startURL

	schan := make(chan []string)
	ichan := make(chan []string)
	fchan := make(chan []Fetched)
	ctx, cancel := context.WithCancel(context.Background())
	// init
	go func() {
		schan <- []string{u}
		c.depth--
	}()
	if c.fileOut != "" && len(c.svFmts) > 0 {
		go fileWorker(c.fileOut, ichan)
	}
	go c.startCheckers(cancel, fchan, ichan, schan)
	c.startFetchers(ctx, schan, fchan)
	endTime := time.Now()
	fmt.Printf("[INFO] Started at %v\n[INFO] Ended at %v\n", beginTime, endTime)
}

func NewCrawler(begin, fo, svFmts, ignr string, cdepth, nc, pc int) *Crawler {
	c := &Crawler{
		startURL: begin,
		depth:    cdepth,
		visited:  make(map[string]bool),
		locker:   make(chan bool, 1),
		netCoro:  nc,
		parsCoro: pc,
	}
	n, err := url.Parse(begin)
	if err != nil {
		log.Fatal(err)
	}
	c.fileOut = fo
	c.svFmts = strings.Split(svFmts, ",")
	c.ignored = strings.Split(ignr, ",")
	c.mainURL = n.Scheme + "://" + n.Host
	return c
}

func init() {
	flag.StringVar(&begin, "url", "", "provides url to crawl")
	flag.StringVar(&outFile, "outfile", "", "provides output file ")
	flag.StringVar(&formats, "formats", "", "comma separated list of `formats`. urls ending with them will be saved into outfile")
	flag.StringVar(&ignore, "ignore", "", "comma separated list of ignored domains")
	flag.IntVar(&dpth, "depth", 1, "provides crawling depth")
	flag.IntVar(&parsRoutines, "parser_threads", 1, "set number of url parsers")
	flag.IntVar(&netRoutines, "network_threads", 1, "set number of data fetchers")
	flag.Parse()
}

func main() {
	fmt.Println("[*] Starting crawler...")
	if begin != "" {
		c := NewCrawler(begin, outFile, formats, ignore, dpth, netRoutines, parsRoutines)
		c.Crawl()
	} else {
		flag.PrintDefaults()
	}
	fmt.Println("[*] Crawler done!")
}
