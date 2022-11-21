# Scraping Handler

![Build Status](https://github.com/vbetsun/scraping/workflows/CI/badge.svg)
[![coverage](https://codecov.io/gh/vbetsun/scraping/branch/master/graph/badge.svg)](https://codecov.io/gh/vbetsun/scraping)
[![GoReport](https://goreportcard.com/badge/github.com/vbetsun/scraping)](https://goreportcard.com/report/github.com/vbetsun/scraping)
![license](https://img.shields.io/github/license/vbetsun/scraping)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/vbetsun/scraping.svg)](https://github.com/vbetsun/scraping)
[![GoDoc](https://pkg.go.dev/badge/github.com/vbetsun/scraping)](https://pkg.go.dev/github.com/vbetsun/scraping)


## Install

```sh
# Go 1.16+
go install github.com/vbetsun/scraping@latest

# Go version < 1.16
go get -u github.com/vbetsun/scraping 
```

## Usage

Run program


```golang
package main

import (
	"log"
	"net/http"

	"github.com/vbetsun/scraping"
)

func main() {
	http.Handle("/", scraping.Handler())
	log.Fatal(http.ListenAndServe(":3000", nil))
}
```

Send request

```sh
curl --location --request POST 'http://localhost:3000' \
--header 'Content-Type: text/plain' \
--data-raw 'https://google.com
https://example.com'
```

Example of response

>16798
>
>1256

## License

Golang Scraping Handler is provided under the [MIT License](LICENSE)