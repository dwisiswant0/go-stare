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
	URL                  *bufio.Scanner
	Verbose              bool
}

// New to proceed screenshots
func New(cfg *Config) {
	var wg sync.WaitGroup
	jobs := make(chan string)

	browserCtx, cancelBrowser := chromedp.NewContext(context.Background())
	defer cancelBrowser()

	// Allocate the browser up-front so that the worker contexts below
	// inherit a single browser instance instead of spinning up their own.
	if err := chromedp.Run(browserCtx); err != nil {
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "[%s] Failed to start browser: %s\n", aurora.Red("!"), err.Error())
		}
		return
	}

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for url := range jobs {
				// Use a fresh tab per URL so that concurrent captures don't
				// interfere with each other and each capture starts from a
				// clean page.
				ctx, cancel := chromedp.NewContext(browserCtx)
				cfg.exec(ctx, url)
				cancel()
			}
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

func (cfg *Config) exec(ctx context.Context, url string) {
	// Apply the timeout per-URL so that a slow target doesn't kill the
	// shared browser context (and every other pending capture) with it.
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, screenshot(url, cfg.Quality, &buf)); err != nil {
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", aurora.Red("!"), err.Error())
		}
		return
	}

	out := path.Join(cfg.Output, replacer(url)+".png")
	if err := ioutil.WriteFile(out, buf, 0644); err != nil {
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "[%s] Failed to create output screenshot: %s\n", aurora.Red("!"), err.Error())
		}
		return
	}
	fmt.Printf("[%s] Screenshot taken for %s (Output: %s)\n", aurora.Green("+"), aurora.Green(url), aurora.Magenta(out))
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
