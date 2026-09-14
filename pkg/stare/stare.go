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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/gologger"
)

type Resolution struct {
	Width  int64
	Height int64
}

type Resolutions []Resolution

func (r *Resolutions) String() string {
	var values []string

	for _, res := range *r {
		values = append(values, fmt.Sprintf("%dx%d", res.Width, res.Height))
	}

	return strings.Join(values, ",")
}

func (r *Resolutions) Set(value string) error {
	parts := strings.Split(value, "x")
	if len(parts) != 2 {
		return fmt.Errorf("invalid resolution %q, expected WIDTHxHEIGHT", value)
	}

	width, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid width %q", parts[0])
	}

	height, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid height %q", parts[1])
	}

	if width <= 0 || height <= 0 {
		return fmt.Errorf("resolution must be positive")
	}

	*r = append(*r, Resolution{
		Width:  width,
		Height: height,
	})

	return nil
}

// Config declare its options
type Config struct {
	Concurrency, Timeout int
	Quality              int64
	Target, Output       string
	Resolutions          Resolutions
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
	err := chromedp.Run(cfg.Context,
		screenshot(
			url,
			cfg.Quality,
			cfg.Resolutions,
			cfg.Output,
			cfg.Verbose,
		),
	)

	if err != nil && cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", aurora.Red("!"), err)
	}
}

func screenshot(
	url string,
	quality int64,
	resolutions Resolutions,
	output string,
	verbose bool,
) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			// if no resolution provided, then use the default that comes with the page
			if len(resolutions) == 0 {
				resolutions = Resolutions{
					{
						Width:  int64(math.Ceil(contentSize.Width)),
						Height: int64(math.Ceil(contentSize.Height)),
					},
				}
			}

			for _, resolution := range resolutions {
				height := resolution.Height

				// override height to capture top to bottom
				if contentSize.Height > float64(height) {
					height = int64(math.Ceil(contentSize.Height))
				}

				err := emulation.SetDeviceMetricsOverride(
					resolution.Width,
					height,
					1,
					false,
				).WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).Do(ctx)

				if err != nil {
					return err
				}

				res, err := page.CaptureScreenshot().
					WithQuality(quality).
					WithClip(&page.Viewport{
						X:      0,
						Y:      0,
						Width:  float64(resolution.Width),
						Height: float64(height),
						Scale:  1,
					}).
					Do(ctx)

				if err != nil {
					return err
				}

				out := path.Join(
					output,
					fmt.Sprintf(
						"%s-%dx%d.png",
						replacer(url),
						resolution.Width,
						resolution.Height,
					),
				)

				if err := ioutil.WriteFile(out, res, 0644); err != nil {
					if verbose {
						fmt.Fprintf(
							os.Stderr,
							"[%s] Failed to create output screenshot: %s\n",
							aurora.Red("!"),
							err,
						)
					}
					continue
				}

				fmt.Printf(
					"[%s] Screenshot taken for %s (Output: %s)\n",
					aurora.Green("+"),
					aurora.Green(url),
					aurora.Magenta(out),
				)
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
