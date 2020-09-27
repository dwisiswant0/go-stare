# go-stare

[![made-with-Go](https://img.shields.io/badge/made%20with-Go-brightgreen.svg)](http://golang.org)
[![go-report](https://goreportcard.com/badge/github.com/dwisiswant0/go-stare?_=1)](https://goreportcard.com/report/github.com/dwisiswant0/go-stare)
[![license](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/dwisiswant0/go-stare/issues)
[![godoc](https://img.shields.io/badge/godoc-reference-brightgreen.svg)](https://godoc.org/github.com/dwisiswant0/go-stare)

A fast & light web screenshot _without_ headless browser but Chrome DevTools Protocol!

<img src="https://user-images.githubusercontent.com/25837540/94014291-86398780-fdd5-11ea-803d-4eb3ec64bd7b.png" height="350">

---

## Resources

- [Installation](#installation)
	- [from Binary](#from-binary)
	- [from Source](#from-source)
	- [from GitHub](#from-github)
- [Usage](#usage)
	- [Basic Usage](#basic-usage)
	- [Flags](#flags)
	- [Target](#target)
		- [Single URL](#single-url)
		- [URLs from list](#urls-from-list)
		- [from Stdin](#from-stdin)
- [Help & Bugs](#help--bugs)
- [License](#license)
- [Version](#version)

## Installation

### from Binary

The installation is easy. You can download a prebuilt binary from [releases page](https://github.com/dwisiswant0/go-stare/releases), unpack and run! or with

```bash
▶ curl -sSfL https://git.io/go-stare | sh -s -- -b /usr/local/bin
```

### from Source

If you have go1.14+ compiler installed and configured:

```bash
▶ GO111MODULE=on go get -v github.com/dwisiswant0/go-stare
```

In order to update the tool, you can use `-u` flag with go get command.

### from GitHub

```bash
▶ git clone https://github.com/dwisiswant0/go-stare
▶ cd go-stare
▶ go build .
▶ mv go-stare /usr/local/bin
```

## Usage

### Basic Usage

Simply, go-stare can be run with:

```bash
▶ go-stare -t "http://domain.tld"
```

### Flags

```bash
▶ go-stare -h
```

This will display help for the tool. Here are all the switches it supports.

| **Flag**          	| **Description**                                               |
|-------------------	|-----------------------------------------------------------    |
| -t, --target      	| Target to captures _(single target URL or list)_              |
| -c, --concurrency 	| Set the concurrency level _(default: 5)_                      |
| -o, --output      	| Screenshot directory output results _(default: ./out)_        |
| -q, --quality     	| Image quality to produce _(default: 75)_                      |
| -T, --timeout     	| Maximum time (seconds) allowed for connection _(default: 30)_ |
| -v, --verbose     	| Verbose mode                                                  |

### Target

You can define a target in 3 ways:

#### Single URL

```bash
▶ go-stare -t "http://domain.tld"
```

#### URLs from list

```bash
▶ go-stare -t /path/to/urls.txt
```

#### from Stdin

In case you want to chained with other tools.

```bash
▶ subfinder -d domain.tld -silent | httpx -silent | go-stare -o ./out
# or
▶ gau domain.tld | go-stare -o ./out -q 50
```

## Help & Bugs

If you are still confused or found a bug, please [open the issue](https://github.com/dwisiswant0/go-stare/issues). All bug reports are appreciated, some features have not been tested yet due to lack of free time.

## License

**go-stare** released under MIT. See `LICENSE` for more details.

## Version

**Current version is 0.0.3-dev** and still development.
