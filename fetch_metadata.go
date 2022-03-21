package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func buildRequest(url, token string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	return req
}

func buildUrl(repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s", repo)
}

func splitLine(s, sep string) (string, string) {
    x := strings.Split(s, sep)
    return x[1], x[2]
}

func createFile(prefix, repo string) (*os.File, error) {
	filePath := path.Join(prefix, repo, "repoMetadata.json")
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	
	return f, nil
}

func createDir(repo string) string {
	dir := fmt.Sprintf("dockerfiles/%s", repo,)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	return dir
}

func writeMetadata(outputPrefix, repo string, content io.ReadCloser) error {
	f, err := createFile(outputPrefix, repo)
	if err != nil {
		return err
	}
	defer f.Close()

	body, err := ioutil.ReadAll(content)
	if err != nil {
		return err
	}

	_, err = f.Write(body)
	if err != nil {
		return err
	}

	return nil
}

func fetchMetadata(repo, outputPrefix string) string {
	url := buildUrl(repo)
	req := buildRequest(url, os.Getenv("GH_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode == 200 {
		err = writeMetadata(outputPrefix, repo, resp.Body)
		if err != nil {
			return err.Error()
		}
	}
	
	return resp.Status
}

func readFile(filePath string) ([]string, int) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, len(lines)
}

func main()  {
	fileName := os.Args[1]
	outputPrefix := os.Args[2]
	repos, repoCount := readFile(fileName)

	for idx, repo := range repos {
		status := fetchMetadata(repo, outputPrefix)
		log.Printf("progress: %d/%d, code: %s, repo: %s \n", idx+1, repoCount, status, repo)
	}
}
