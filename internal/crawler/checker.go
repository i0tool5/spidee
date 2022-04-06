package crawler

import (
	"fmt"
	"log"

	"github.com/i0tool5/spidee/internal/core"
	"github.com/i0tool5/spidee/internal/misc"
)

func (c *Crawler) genSlices(baseAddr string, hrefs []string) (
	urls, tout []string) {

	urls = make([]string, 0)
	tout = make([]string, 0)

	for _, href := range hrefs {
		addr, err := misc.GenAddr(baseAddr, href)
		if err != nil {
			log.Printf("[!] error occured %v\n", err)
			break
		}

		if !c.visit(addr) {
			urls = append(urls, addr)
			// fmt.Printf("[C%d][Found]> %s\n", n, addr)
			for _, ending := range c.constraints.SaveFmts {
				if misc.EndsWith(addr, ending) {
					tout = append(tout, addr)
				}
			}
		}
	}

	return
}

func (c *Crawler) checker(fetchArr chan core.FetchedArr, tf, o chan []string) {
	for {
		fetched, ok := <-fetchArr
		if !ok || len(fetched) < 1 {
			log.Println("[HANDLER] nothing to do")
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
			// fmt.Println("[DEBUG] base", baseAddr)
			// fmt.Println("[DEBUG] hrefs", hrefs)

			urls, tout = c.genSlices(baseAddr, hrefs)

			if len(c.constraints.SaveFmts) > 0 && c.cfg.FileOut != "" {
				tf <- tout
			}
			fmt.Println("len:", len(urls), len(tout))

			for _, url := range tout {
				log.Printf("got url: %v\n", url)
			}

			fmt.Println("Depth:", c.cfg.Depth)
			if c.cfg.Depth <= 0 {
				log.Println("[HANDLER] depth exceeded")
				c.wg.Done()
				return
			}
			o <- urls
			c.cfg.Depth-- // TODO: make it atomic since it can run in a goroutine
		}
	}
}

func (c *Crawler) startCheckers(cancelFunc func(), fetchArrCh chan core.FetchedArr,
	fout, out chan []string) {

	for i := 0; i < c.cfg.ParsGoro; i++ {
		c.wg.Add(1)
		go c.checker(fetchArrCh, fout, out)
	}

	c.wg.Wait()
	fmt.Println("[checkers] done")
	cancelFunc()
	close(fout)
	close(out)
}
