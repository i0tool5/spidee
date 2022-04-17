package crawler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/i0tool5/spidee/internal/core"
	"github.com/i0tool5/spidee/internal/misc"
)

// Config for the crawler
type Config struct {
	Depth    int
	NetGoro  int
	ParsGoro int
	// mainURL  string
	StartURL string
	FileOut  string
}

// Constraints for crawler
type Constraints struct {
	SaveFmts []string
	Ignored  []string
}

// Crawler is a main struct for crawling web sites
type Crawler struct {
	visited     map[string]bool
	mutex       *sync.Mutex
	depthMut    *sync.Mutex
	wg          *sync.WaitGroup
	cfg         *Config
	constraints *Constraints
}

func (c *Crawler) send(urls []string, ch chan string) {
	for _, url := range urls {
		if !c.isIgnored(url) {
			ch <- url
		}
	}
	close(ch)
}

func (c *Crawler) fetch(wg *sync.WaitGroup, urlCh chan string, out chan core.Fetched) {
	for url := range urlCh {
		fmt.Printf("[Fetching]> %s\n", url)
		dat, err := core.Fetch(url)
		if err != nil {
			log.Printf("error %v\n", err)
			continue
		}
		out <- *dat
	}
	wg.Done()
}

// TODO: rename
func (c *Crawler) blah(urls []string, goros int) (fetched core.FetchedArr) {
	fetched = make(core.FetchedArr, 0)
	var (
		wg    = new(sync.WaitGroup)
		urlCh = make(chan string)
		outCh = make(chan core.Fetched)
		d     = make(chan struct{}, 1)
	)

	// filling channel with urls and sends it to `fetcher`
	go c.send(urls, urlCh)

	wg.Add(goros)
	for i := 0; i < goros; i++ {
		go c.fetch(wg, urlCh, outCh)
	}

	go func() {
		for {
			v, ok := <-outCh
			if !ok {
				d <- struct{}{}
				return
			}
			fetched = append(fetched, v)
		}
	}()

	wg.Wait()
	close(outCh)
	<-d
	return
}

func (c *Crawler) startFetchers(ctx context.Context, urlsCh chan []string,
	fetchedArrCh chan core.FetchedArr) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[FETCHERS] Signal Done.")
			close(fetchedArrCh)
			return
		case urls, ok := <-urlsCh:
			if !ok || len(urls) < 1 {
				close(fetchedArrCh)
				return
			}

			z := c.blah(urls, c.cfg.NetGoro)
			fetchedArrCh <- z
		}
	}
}

func (c *Crawler) isIgnored(url string) bool {
	for _, ig := range c.constraints.Ignored {
		if strings.Contains(url, ig) && ig != "" {
			return true
		}
	}
	return false
}

func (c *Crawler) visit(url string) (resp bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.visited[url] {
		resp = true
	}
	c.visited[url] = true
	return
}

// func (c *Crawler) ChangeDepth() {
// 	atomic.AddUint64()
// }

func (c *Crawler) Crawl() {

	var (
		beginTime   = time.Now()
		startURL    = c.cfg.StartURL
		urlsChan    = make(chan []string)
		fileChan    = make(chan []string)
		fetchArrCh  = make(chan core.FetchedArr)
		ctx, cancel = context.WithCancel(context.Background())
	)

	// init
	go func() {
		urlsChan <- []string{startURL}
		c.cfg.Depth--
	}()

	if c.cfg.FileOut != "" && len(c.constraints.SaveFmts) > 0 {
		go misc.FileWorker(c.cfg.FileOut, fileChan)
	}

	// TODO: refactor
	go c.startCheckers(cancel, fetchArrCh, fileChan, urlsChan)
	c.startFetchers(ctx, urlsChan, fetchArrCh)

	endTime := time.Now()
	fmt.Printf("[INFO] Started at %v\n[INFO] Ended at %v\n", beginTime, endTime)
}

// NewCrawler returns new crawler instance
func NewCrawler(cfg *Config, constraints *Constraints /*,fetcher interface*/) (c *Crawler) {
	c = &Crawler{
		cfg:         cfg,
		constraints: constraints,
		wg:          new(sync.WaitGroup),
		visited:     make(map[string]bool),
		mutex:       new(sync.Mutex),
		depthMut:    new(sync.Mutex),
	}

	return
}

// func NewCrawler(begin, fo, svFmts, ignr string, cdepth, nc, pc int) *Crawler {
// 	c := &Crawler{
// 		startURL: begin,
// 		depth:    cdepth,
// 		visited:  make(map[string]bool),
// 		locker:   make(chan bool, 1),
// 		netCoro:  nc,
// 		parsCoro: pc,
// 	}
// 	n, err := url.Parse(begin)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	c.fileOut = fo
// 	c.svFmts = strings.Split(svFmts, ",")
// 	c.ignored = strings.Split(ignr, ",")
// 	c.mainURL = n.Scheme + "://" + n.Host
// 	return c
// }
