package main

import (
	"log"
	"net/http"

	"github.com/vbetsun/scraping"
)

func main() {
	http.Handle("/custom-path", scraping.Handler())
	log.Fatal(http.ListenAndServe(":3000", nil))
}
