package main

import (
	"flag"
	"fmt"

	"github.com/i0tool5/spidee/internal/core/crawler"
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

func init() {
	flag.StringVar(&begin, "url", "", "provides url to crawl")
	flag.StringVar(&outFile, "outfile", "", "provides output file ")
	flag.StringVar(&formats, "formats", "", "comma separated list of `formats`. urls ending with them will be saved into outfile")
	flag.StringVar(&ignore, "ignore", "", "comma separated list of ignored domains")
	flag.IntVar(&dpth, "depth", 1, "provides crawling depth")
	flag.IntVar(&parsRoutines, "parsers", 1, "Set number of url parser goroutines")
	flag.IntVar(&netRoutines, "fetchers", 1, "Set number of data fetcher goroutines")
	flag.Parse()
}

func main() {
	fmt.Println("[*] Starting crawler...")
	if begin != "" {
		c := crawler.NewCrawler(begin, outFile, formats, ignore, dpth, netRoutines, parsRoutines)
		c.Crawl()
	} else {
		flag.PrintDefaults()
	}
	fmt.Println("[*] Crawler done!")
}
