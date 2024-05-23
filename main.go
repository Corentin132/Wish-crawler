package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const wishCrawler = `  _       __   _            __                                                  __               
| |     / /  (_)  _____   / /_           _____   _____   ____ _   _      __   / /  ___     _____
| | /| / /  / /  / ___/  / __ \         / ___/  / ___/  / __ '/  | | /| / /  / /  / _ \   / ___/
| |/ |/ /  / /  (__  )  / / / /        / /__   / /     / /_/ /   | |/ |/ /  / /  /  __/  / /    
|__/|__/  /_/  /____/  /_/ /_/         \___/  /_/      \__,_/    |__/|__/  /_/   \___/  /_/     
                                                                                                `

// Made for learner how Web Crawler works,
// TODO:: add filter by size
const maxWorkers = 10 // max goroutine workers put less to not be detected maybe 10 is too much

func getContentLength(resp *http.Response) int64 {
	if resp.ContentLength != -1 {
		return resp.ContentLength
	}

	buf := make([]byte, 32*1024) // 32 KB buffer
	var total int64
	for {
		n, err := resp.Body.Read(buf)
		total += int64(n)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Handle other errors if needed
			return -1
		}
	}
	return total
}

func Crawl(wordList string, target string, recursion bool, sValues []int, depth int) {
	file, err := os.Open(wordList)
	if err != nil {
		fmt.Println("Error opening word list: ", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	jobs := make(chan string)

	// Start worker goroutines
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg, recursion, wordList, sValues, depth)
	}

	// Read lines from the file and send them to the jobs channel
	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				jobs <- fmt.Sprintf("%s/%s", target, line)
			}
		}
		close(jobs)
	}()

	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading word list: ", err)
	}
}

func worker(jobs <-chan string, wg *sync.WaitGroup, recursion bool, wordlist string, sValues []int, depth int) {
	defer wg.Done()
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	indent := strings.Repeat("  ", depth) // Indentation based on depth

	for word := range jobs {
		resp, err := client.Get(word)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		for _, code := range sValues {
			if code == resp.StatusCode {
				fmt.Println("result body : ", getContentLength(resp))
				fmt.Printf("%sFound: %s for status code: %d\n", indent, word, code)
				if recursion {
					fmt.Printf("%sRecursion: %s\n", indent, word)
					Crawl(wordlist, word, recursion, sValues, depth+1)
				}
				break
			}
		}
	}
}

func main() {
	target := flag.String("t", "https://google.com", "Target URL")
	wordList := flag.String("w", "/path/to/list", "Path to word list")
	recursion := flag.Bool("r", false, "Enable recursion")
	status := flag.String("s", "", "Space-separated list of status codes")

	flag.Parse()

	var sValues []int
	if *status != "" {
		for _, str := range strings.Split(*status, " ") {
			val, err := strconv.Atoi(str)
			if err != nil {
				fmt.Printf("Error converting '%s' to integer: %v\n", str, err)
				return
			}
			sValues = append(sValues, val)
		}
	}

	fmt.Println(wishCrawler)
	fmt.Printf("You chose this option: %s | %s | %s \n", *target, *wordList, *status)
	if *wordList == "" {
		fmt.Println("Please provide a word list in a flag.")
		return
	}

	Crawl(*wordList, *target, *recursion, sValues, 0)
}
