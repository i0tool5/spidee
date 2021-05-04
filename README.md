# spidee
Simple crawler written in Go.

## Known issues and problems
- The user must specify both the format and the file arguments to save what crawler finds to a file. 
***I will fix this later.***

## Purpose
I wrote this simple web-crawler for learning purposes... and for fun!

## Usage
```golang
  -depth
        provides crawling depth (default 1)
  -fetchers
        Set number of data fetcher goroutines (default 1)
  -formats
        comma separated list of formats. urls ending with them will be saved into outfile
  -ignore
        comma separated list of ignored domains
  -outfile
        provides output file
  -parsers
        Set number of url parser goroutines (default 1)
  -url
        provides url to crawl
```
