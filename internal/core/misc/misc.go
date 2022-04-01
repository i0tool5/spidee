package misc

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

// FileWorker is a function to handle output file
func FileWorker(fileName string, names chan []string) {
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

// EndsWith is a simple wrapper around strings.HasSuffix
func EndsWith(s, ending string) bool {
	return strings.HasSuffix(s, ending)
}
