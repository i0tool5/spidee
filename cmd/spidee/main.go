package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/i0tool5/spidee/internal/crawler"
)

// CLI arguments variables
var (
	depth        int
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
	flag.IntVar(&depth, "depth", 1, "provides crawling depth")
	flag.IntVar(&parsRoutines, "parsers", 1, "Set number of url parser goroutines")
	flag.IntVar(&netRoutines, "fetchers", 1, "Set number of data fetcher goroutines")
	flag.Parse()
}

func main() {
	if begin == "" {
		flag.PrintDefaults()
		return
	}

	fmt.Println("[*] Starting crawler...")

	cfg := &crawler.Config{
		Depth:    int32(depth),
		NetGoro:  netRoutines,
		ParsGoro: parsRoutines,
		StartURL: begin,
		FileOut:  outFile,
	}

	constraints := &crawler.Constraints{
		SaveFmts: strings.Split(formats, ","),
		Ignored:  strings.Split(ignore, ","),
	}

	c := crawler.NewCrawler(cfg, constraints)
	c.Crawl()

	fmt.Println("[*] Crawler done!")
}
