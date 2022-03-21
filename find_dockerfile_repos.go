package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// filter duplicates from a list based on a second list
func filter(list []string, filter []string) []string {
	var filtered []string
	for _, item := range list {
		if !contains(filter, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// check if a list contains a string
func contains(list []string, item string) bool {
	for _, val := range list {
		if val == item {
			return true
		}
	}
	return false
}

// return first n item from a list skip offset
func slice(list []string, n int, offset int) []string {
	var sliced []string
	for i, item := range list {
		if i >= offset && i < offset+n {
			sliced = append(sliced, item)
		}
	}
	return sliced
}

func writeFile(visitedPath, line string) {
	file, err := os.OpenFile(visitedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if _, err := file.WriteString(line + "\n"); err != nil {
		log.Fatalln(err)
	}
}

func buildRequest(url, token string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	return req
}

func checkRepo(repo, token, outputPath, visitedPath, statusPath, failedPath string) {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/languages", repo)
	resp, err := client.Do(buildRequest(url, token))
	if err != nil {
		fmt.Println(repo, err)
	}
	defer resp.Body.Close()

	log.Println(fmt.Sprintf("checking %s - %s", repo, resp.Status))
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		langs := make(map[string]string)
		json.Unmarshal(body, &langs)
		if _, ok := langs["Dockerfile"]; ok {
			writeFile(outputPath, repo)
		}
		writeFile(visitedPath, repo)
	} else if resp.StatusCode == 404 {
		writeFile(visitedPath, repo)
	} else {
		writeFile(failedPath, repo)
	}
	writeFile(statusPath, fmt.Sprintf("%d - %s", resp.StatusCode, repo))
}

func getRateLimit(token string) int {
	client := &http.Client{}
	resp, err := client.Do(buildRequest("https://api.github.com/rate_limit", token))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer resp.Body.Close()

	remaining, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	return remaining
}

func readFile(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func getTokens(filePath string) []string {
	tokens := readFile(filePath)

	return tokens
}

func main() {
	tokens := getTokens("../.gh_tokens")
	repos := readFile("./data/engineered_projects_failed.txt")
	visited := readFile("./data/engineered_projects_visited.txt")
	filtered := filter(repos, visited)
	fmt.Println(fmt.Sprintf("repos %d", len(repos)))
	fmt.Println(fmt.Sprintf("visited %d", len(visited)))
	fmt.Println(fmt.Sprintf("filtered %d", len(filtered)))

	offset := 0
	for {
		var wg sync.WaitGroup
		for _, token := range tokens {
			remaining := getRateLimit(token)
			if remaining > 0 {
				toCheck := slice(filtered, remaining, offset)
				offset += len(toCheck)
				fmt.Println(fmt.Sprintf("%s - toCheck: %d, offset: %d", token, len(toCheck), offset))

				wg.Add(1)
				go func(toCheck []string, token string) {
					if len(toCheck) != 0 {
						for _, repo := range toCheck {
							checkRepo(
								repo,
								token,
								"./data/engineered_projects_confirmed.txt",
								"./data/engineered_projects_visited.txt",
								"./data/engineered_projects_status.txt",
								"./data/engineered_projects_failed.txt",
							)
						}
					} else {
						fmt.Println("No more repos to check")
					}

					defer wg.Done()
				}(toCheck, token)
			}
		}
		wg.Wait()
		fmt.Println("Goroutines finished, sleeping for 61 min")
		time.Sleep(time.Minute * 61)
	}
}
