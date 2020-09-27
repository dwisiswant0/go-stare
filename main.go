package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/dwisiswant0/go-stare/internal"
	"github.com/dwisiswant0/go-stare/pkg/stare"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/gologger"
)

const (
	author  = "dwisiswant0"
	version = "0.0.1"
	banner  = `
                      _
   __ _  ___ ____ ___| |_ __ _ _ __ ___ 
  / _' |/ _ \____/ __| __/ _' | '__/ _ \
 | (_| | (_) |   \__ \ || (_| | | |  __/
  \__, |\___/    |___/\__\__,_|_|  \___|
   __/ |
  |___/     v` + version + ` - @` + author + `

`
)

var cfg *stare.Config

func init() {
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu + 1)

	cfg = &stare.Config{}

	flag.StringVar(&cfg.Target, "target", "", "")
	flag.StringVar(&cfg.Target, "t", "", "")

	flag.IntVar(&cfg.Concurrency, "concurrency", 5, "")
	flag.IntVar(&cfg.Concurrency, "c", 5, "")

	flag.StringVar(&cfg.Output, "output", "./out", "")
	flag.StringVar(&cfg.Output, "o", "./out", "")

	flag.Int64Var(&cfg.Quality, "quality", 75, "")
	flag.Int64Var(&cfg.Quality, "q", 75, "")

	flag.IntVar(&cfg.Timeout, "timeout", 10, "")
	flag.IntVar(&cfg.Timeout, "T", 10, "")

	flag.BoolVar(&cfg.Verbose, "verbose", false, "")
	flag.BoolVar(&cfg.Verbose, "v", false, "")

	fmt.Fprintf(os.Stderr, "%s", aurora.Bold(aurora.Cyan(banner)))

	flag.Usage = func() {
		h := "A fast & light web screenshot without headless browser but Chrome DevTools Protocol!\n\n"

		h += "Usage:\n"
		h += "  go-stare -t [URL|URLs.txt] -o [outputDir]\n\n"

		h += "Options:\n"
		h += "  -t, --target <URL/FILE>     Target to captures (single target URL or list)\n"
		h += "  -c, --concurrency <int>     Set the concurrency level (default: 50)\n"
		h += "  -o, --output <DIR>          Screenshot directory output results (default: ./out)\n"
		h += "  -q, --quality <int>         Image quality to produce (default: 75)\n"
		h += "  -T, --timeout <int>         Maximum time (seconds) allowed for connection (default: 10)\n"
		h += "  -v, --verbose               Verbose mode\n\n"

		h += "Examples:\n"
		h += "  go-stare -t http://domain.tld\n"
		h += "  go-stare -t urls.txt -o ./out\n"
		h += "  cat urls.txt | go-stare -o ./out -q 90\n\n"

		fmt.Fprint(os.Stderr, h)
	}

	flag.Parse()
}

func main() {
	internal.Validator(cfg)

	if err := os.MkdirAll(cfg.Output, 0750); err != nil {
		gologger.Fatalf("Failed to create output directory: %s", err.Error())
		os.Exit(1)
	}

	stare.New(cfg)
}
