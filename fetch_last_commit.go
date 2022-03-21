package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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

func writeContent(filePath string, body []byte) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	_, err = f.Write(body)
	if err != nil {
		log.Fatalln(err)
	}
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

func splitLine(s, sep string) []string {
  return strings.Split(s, sep)
}

func buildRequest(url, token string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	return req
}

func buildUrl(line string) string {
	data := splitLine(line, ",")
	file := strings.Replace(data[2], " ", "%20", -1)
	return fmt.Sprintf("https://api.github.com/repos/%s/commits?path=%s&page=1&per_page=1", data[1], file)
}

func checkFileExists(prefix, line string) (string, bool) {
	data := splitLine(line, ",")
	fmt.Println(data)
	filePath := path.Join(prefix, data[1], strings.Replace(data[2], "/", "", -1))
	_, err := os.Stat(filePath)
	return filePath, err == nil
}

func getLastCommit(prefix, line, token, visitedPath, statusPath, failedPath string) {
	filePath, ok := checkFileExists(prefix, line)
	if !ok {
		return
	}

	url := buildUrl(line)
	repo := splitLine(line, ",")[1]
	client := &http.Client{}
	resp, err := client.Do(buildRequest(url, token))
	if err != nil {
		fmt.Println(repo, err)
	}
	defer resp.Body.Close()

	log.Println(fmt.Sprintf("getting last commit %s - %s", repo, resp.Status))
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		writeContent(filePath + ".lastCommit.json", body)
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
	prefix := os.Args[1]
	tokens := getTokens("../.gh_tokens")
	dockerfiles := readFile("./last_commit/dockerfiles_paths_failed.txt")
	// dockerfiles := readFile("./last_commit/dockerfiles_paths.txt")
	visited := readFile("./last_commit/dockerfiles_paths_visited.txt")
	filtered := filter(dockerfiles, visited)
	fmt.Println(fmt.Sprintf("repos %d", len(dockerfiles)))
	fmt.Println(fmt.Sprintf("visited %d", len(visited)))
	fmt.Println(fmt.Sprintf("filtered %d", len(filtered)))

	offset := 0
	for {
		var wg sync.WaitGroup
		for _, token := range tokens {
			remaining := getRateLimit(token)
			if remaining > 0 {
				toCheck := slice(filtered, remaining-1500, offset)
				offset += len(toCheck)
				fmt.Println(fmt.Sprintf("%s - toCheck: %d, offset: %d", token, len(toCheck), offset))

				wg.Add(1)
				go func(toCheck []string, token string) {
					if len(toCheck) != 0 {
						for _, line := range toCheck {
							getLastCommit(
								prefix,
								line,
								token,
								"./last_commit/dockerfiles_paths_visited.txt",
								"./last_commit/dockerfiles_paths_status.txt",
								"./last_commit/dockerfiles_paths_failed.txt",
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
