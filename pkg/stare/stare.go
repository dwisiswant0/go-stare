package stare

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/gologger"
)

// Config declare its options
type Config struct {
	Concurrency, Timeout int
	Quality              int64
	Target, Output       string
	Buffer               []byte
	URL                  *bufio.Scanner
	Verbose              bool
	Context              context.Context
	CtxCancel            context.CancelFunc
}

// New to proceed screenshots
func New(cfg *Config) {
	var wg sync.WaitGroup
	jobs := make(chan string)

	cfg.Context, _ = chromedp.NewContext(context.Background())
	cfg.Context, cfg.CtxCancel = context.WithTimeout(cfg.Context, time.Duration(cfg.Timeout)*time.Second)
	defer cfg.CtxCancel()

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			for url := range jobs {
				cfg.exec(url)
			}
			defer wg.Done()
		}()
	}

	for cfg.URL.Scan() {
		u := cfg.URL.Text()
		if isURL(u) {
			jobs <- u
		} else {
			if cfg.Verbose {
				fmt.Fprintf(os.Stderr, "[%s] Invalid URL of %s\n", aurora.Red("!"), aurora.Red(u))
			}
		}
	}

	close(jobs)
	wg.Wait()
}

func (cfg *Config) exec(url string) {
	if err := chromedp.Run(cfg.Context, screenshot(url, cfg.Quality, &cfg.Buffer)); err != nil {
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", aurora.Red("!"), err.Error())
		}
	}

	out := path.Join(cfg.Output, replacer(url)+".png")
	if err := ioutil.WriteFile(out, cfg.Buffer, 0644); err != nil {
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "[%s] Failed to create output screenshot: %s\n", aurora.Red("!"), err.Error())
		}
	} else {
		fmt.Printf("[%s] Screenshot taken for %s (Output: %s)\n", aurora.Green("+"), aurora.Green(url), aurora.Magenta(out))
	}
}

func screenshot(url string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}

			return nil
		}),
	}
}

func replacer(url string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		gologger.Errorf("Failed to replace non-alphas URL: %s", err.Error())
		os.Exit(2)
	}
	return reg.ReplaceAllString(url, "_")
}

func isURL(s string) bool {
	_, e := url.ParseRequestURI(s)
	if e != nil {
		return false
	}

	u, e := url.Parse(s)
	if e != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
