package internal

import (
	"bufio"
	"os"
	"strings"

	"github.com/dwisiswant0/go-stare/pkg/stare"
	"github.com/projectdiscovery/gologger"
)

// Validator to validate options
func Validator(cfg *stare.Config) {
	if isStdin() {
		cfg.URL = bufio.NewScanner(os.Stdin)
	} else if cfg.Target != "" {
		if strings.HasPrefix(cfg.Target, "http") {
			cfg.URL = bufio.NewScanner(strings.NewReader(cfg.Target))
		} else {
			r, err := os.Open(cfg.Target)
			if err != nil {
				gologger.Errorf("Invalid '%s' URL or file!", cfg.Target)
				gologger.Infof("Use -h flag for more info about command.")
				os.Exit(1)
			}
			cfg.URL = bufio.NewScanner(r)
		}
	} else {
		gologger.Errorf("No target inputs provided!")
		gologger.Infof("Use -h flag for more info about command.")
		os.Exit(1)
	}
}

func isStdin() bool {
	f, e := os.Stdin.Stat()
	if e != nil {
		return false
	}

	if f.Mode()&os.ModeNamedPipe == 0 {
		return false
	}

	return true
}
