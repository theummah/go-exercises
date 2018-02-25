package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Visit struct {
	fetched map[string]bool
	mux sync.Mutex
}


// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, c chan string) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	//fmt.Printf("At depth %v.", depth)
	defer close(c)
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		visited_urls.mux.Lock()
		visited_urls.fetched[url] = true
		visited_urls.mux.Unlock()
		
		c <- fmt.Sprintln(err)
		return
	}
	
	visited_urls.mux.Lock()
	visited_urls.fetched[url] = true
	visited_urls.mux.Unlock()
	
	c <- fmt.Sprintf("found: %s %q\n", url, body)
	
	c_1 := make([]chan string, len(urls))
	
	for i, u := range urls {
		visited_urls.mux.Lock()
		if !visited_urls.fetched[u] {
			visited_urls.mux.Unlock()
			c_1[i] = make(chan string)
			go Crawl(u, depth-1, fetcher, c_1[i])
		} else {
			//close(c_1[i])
			visited_urls.mux.Unlock()
		}
	}
	
	for _, s := range c_1 {
		if s != nil {
			c  <- <- s
		}
	}
	//fmt.Printf("Finishing depth %v.\n", depth)
	return
}

func main() {
	c := make(chan string)
	go Crawl("http://golang.org/", 4, fetcher, c)
	
	for s := range c {
		fmt.Println(s)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

var visited_urls = Visit{fetched: make(map[string]bool)}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

