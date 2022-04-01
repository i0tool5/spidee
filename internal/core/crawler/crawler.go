package crawler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/i0tool5/spidee/core"
	"github.com/i0tool5/spidee/core/misc"
)

// Config for the crawler
type Config struct {
	depth    int
	netGoro  int
	parsGoro int
	mainURL  string
	startURL string
	fileOut  string
}

// Constraints for crawler
type Constraints struct {
	svFmts  []string
	ignored []string
}

// Crawler is a main struct for crawling web sites
type Crawler struct {
	locker   	chan bool
	visited  	map[string]bool
	cfg      	Config
	constraints Constraints
}

func (c *Crawler) startFetchers(ctx context.Context, inpCh chan []string, outCh chan core.FetchedArr) {
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
			z := make(core.FetchedArr, 0)
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
						fmt.Printf("[F%d][Fetching]> %s\n", n, url)
						dat := core.Fetch(url)
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

func genAddr(base, href string) (addr string, err error) {
	addr = "/"
	if strings.HasPrefix(href, "http") {
		addr = href
	} else {
		a, err := url.Parse(base)
		if err != nil {
			return
		}
		b, err := url.Parse(href)
		if err != nil {
			return
		}
		resolved := a.ResolveReference(b)
		addr = resolved.String()
	}
	return
}

func (c *Crawler) blah() { //TODO: rename

} 

func (c *Crawler) checker(n int, cf chan core.FetchedArr, tf, o chan []string, d chan bool) {
	for {
		fetched, ok := <-cf
		if !ok {
			d <- true
			return
		}

		urls := make([]string, 0)
		tout := make([]string, 0)

		for _, fStruct := range fetched {
			baseAddr := fStruct.Base()
			hrefs := fStruct.Hrefs()
			for _, href := range hrefs {
				addr, err := genAddr(baseAddr, href)
				if err != nil {
					log.Printf("[!] error occured %v\n", err)
					break
				}

				if !c.visit(addr) {
					urls = append(urls, addr)
					fmt.Printf("[C%d][Found]> %s\n", n, addr)
					for _, ending := range c.constraints.svFmts {
						if misc.EndsWith(addr, ending) {
							tout = append(tout, addr)
						}
					}
				}
			}

			if len(c.constraints.svFmts) > 0 && c.cfg.fileOut != "" {
				tf <- tout
			}
			if c.depth <= 0 {
				fmt.Println("[HANDLER] Depth exceeded!")
				cancelFunc()
				d <- true
				return
			}
			o <- urls
			c.depth-- // TODO: make it atomic since it can run in a goroutine 
		}
	}
}

func (c *Crawler) startCheckers(cancelFunc func(), ch chan core.FetchedArr, fout, out chan []string) {
	done := make(chan bool, c.cfg.parsCoro)
	for i := 0; i < c.cfg.parsCoro; i++ {
		// TODO: replace it with checker func
		go func(n int, cf chan core.FetchedArr, tf, o chan []string, d chan bool) {
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
					hrefs := fStruct.GetHrefs()
					for _, href := range hrefs {
						if strings.HasPrefix(href, "http") {
							addr = href
						} else {
							a, err := url.Parse(fStruct.GetBase())
							if err != nil {
								log.Printf("[!] Error occured parsing url: %s\n", err)
								break
							}
							b, err := url.Parse(href)
							if err != nil {
								log.Printf("[!] Error occured: %s\n", err)
								break
							}
							resolved := a.ResolveReference(b)
							addr = resolved.String()
						}
						if !c.visit(addr) {
							urls = append(urls, addr)
							fmt.Printf("[C%d][Found]> %s\n", n, addr)
							for _, ending := range c.svFmts {
								if misc.EndsWith(addr, ending) {
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

func (c *Crawler) visit(url string) (resp bool) {
	//TODO: use mutex instead of chanels
	c.locker <- true
	if c.visited[url] {
		resp = true
	}
	c.visited[url] = true
	<-c.locker
	return resp
}

// func (c *Crawler) ChangeDepth() {
// 	atomic.AddUint64()
// }

func (c *Crawler) Crawl() {
	beginTime := time.Now()
	u := c.startURL

	schan := make(chan []string)
	ichan := make(chan []string)
	fchan := make(chan core.FetchedArr)
	ctx, cancel := context.WithCancel(context.Background())
	// init
	go func() {
		schan <- []string{u}
		c.depth--
	}()
	if c.fileOut != "" && len(c.svFmts) > 0 {
		go misc.FileWorker(c.fileOut, ichan)
	}
	go c.startCheckers(cancel, fchan, ichan, schan)
	c.startFetchers(ctx, schan, fchan)
	endTime := time.Now()
	fmt.Printf("[INFO] Started at %v\n[INFO] Ended at %v\n", beginTime, endTime)
}

// NewCrawler returns new crawler instance
func NewCrawlerTODO(cfg *Config /*,fetcher interface*/) (c *Crawler, err error) {
	c = &Crawler{
		cfg: cfg,
		visited:  make(map[string]bool),
		locker:   make(chan bool, 1),
	}
	n, err := url.Parse(begin)
	if err != nil {
		return
	}
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