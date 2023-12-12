package crawler

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/i0tool5/spidee/internal/core"
	"github.com/i0tool5/spidee/internal/misc"
)

func (c *Crawler) genSlices(baseAddr string, hrefs []string) (
	urls, tout []string,
) {

	urls = make([]string, 0)
	tout = make([]string, 0)

	for _, href := range hrefs {
		addr, err := misc.GenAddr(baseAddr, href)
		if err != nil {
			log.Printf("[!] error occurred %v\n", err)
			break
		}

		if !c.visit(addr) {
			urls = append(urls, addr)
			for _, ending := range c.constraints.SaveFmts {
				if misc.EndsWith(addr, ending) {
					tout = append(tout, addr)
				}
			}
		}
	}

	return
}

func (c *Crawler) checker(fetchArr <-chan core.FetchedArr, tf, o chan []string) {
	for {
		fetched, ok := <-fetchArr
		if !ok || len(fetched) < 1 {
			c.wg.Done()
			return
		}

		var (
			urls []string
			tout []string
		)

		for _, fStruct := range fetched {
			baseAddr := fStruct.Base()
			hrefs := fStruct.Hrefs()

			urls, tout = c.genSlices(baseAddr, hrefs)

			if len(c.constraints.SaveFmts) > 0 && c.cfg.FileOut != "" {
				tf <- tout
			}

			// DEBUG
			// print urls
			for _, url := range urls {
				log.Printf("got url: %v\n", url)
			}

			if c.cfg.Depth > 0 {
				o <- urls
			}
			atomic.AddInt32(&c.cfg.Depth, -1)
		}
	}
}

func (c *Crawler) startCheckers(cancelFunc func(), fetchArrCh chan core.FetchedArr,
	fout, out chan []string) {

	c.wg.Add(c.cfg.ParsGoro)
	for i := 0; i < c.cfg.ParsGoro; i++ {
		go c.checker(fetchArrCh, fout, out)
	}

	// monitor crawler depth
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if c.cfg.Depth < 1 {
				cancelFunc() // cancel fetchers
				return
			}
		}
	}()

	c.wg.Wait()
	fmt.Println("[checkers] done")
	close(fout)
	close(out)
}
