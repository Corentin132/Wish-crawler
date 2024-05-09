package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const wishCrawler = `  _       __   _            __                                                  __               
| |     / /  (_)  _____   / /_           _____   _____   ____ _   _      __   / /  ___     _____
| | /| / /  / /  / ___/  / __ \         / ___/  / ___/  / __ '/  | | /| / /  / /  / _ \   / ___/
| |/ |/ /  / /  (__  )  / / / /        / /__   / /     / /_/ /   | |/ |/ /  / /  /  __/  / /    
|__/|__/  /_/  /____/  /_/ /_/         \___/  /_/      \__,_/    |__/|__/  /_/   \___/  /_/     
                                                                                                `

// Made for learner how Web Crawler works,
// TODO:: add filter by size of the resp because i get 200 status but the 404 page 🤡
const maxWorkers = 10 // max goroutine workers

func Crawl(wordList string, target string, recursion bool) {
	file, err := os.Open(wordList)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer file.Close()

	r := bufio.NewReader(file)
	var wg sync.WaitGroup
	jobs := make(chan string)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg, recursion, wordList)
	}

	go func() {
		for {
			line, _, err := r.ReadLine()
			if len(line) > 0 {
				jobs <- fmt.Sprintf("%s/%s", target, line)
			}
			if err != nil {
				close(jobs)
				break
			}
		}
	}()

	wg.Wait()
}

func worker(jobs <-chan string, wg *sync.WaitGroup, recursion bool, wordlist string) {
	defer wg.Done()
	for word := range jobs {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := client.Get(word)
		if err != nil {
			continue
		}
		if resp.StatusCode != 404 {
			fmt.Println("Found: ", word)
			if recursion {
				fmt.Println("Recursion: ", word)
				Crawl(wordlist, word, recursion)

			}
		}
	}
}

func main() {
	target := flag.String("t", "https://google.com", "a string")
	wordList := flag.String("w", "/path/to/list", "a string")
	recursion := flag.Bool("r", false, "add recursion to the word list")

	flag.Parse()
	fmt.Println(wishCrawler)
	fmt.Printf("You chose this option: %s | %s \n", *target, *wordList)
	if *wordList == "" {
		fmt.Println("Please give a word list in a flag.")
		return
	}
	Crawl(*wordList, *target, *recursion)

}
