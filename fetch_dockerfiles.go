package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func buildUrl(line string) string {
	repo, file := splitLine(line, ",")
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/master/%s", repo, file)
}

func splitLine(s, sep string) (string, string) {
    x := strings.Split(s, sep)
    return x[1], x[2]
}

func createDockerfile(dir, dockerfile string) *os.File {
	dockerfileName := strings.Replace(dockerfile, "/", "", -1)
	f, err := os.Create(fmt.Sprintf("%s/%s", dir, dockerfileName))
	if err != nil {
		log.Fatalln(err)
	}
	
	return f
}

func createDir(repo string) string {
	dir := fmt.Sprintf("dockerfiles/%s", repo,)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	return dir
}

func writeDockerfile(repo, dockerfile string, content io.ReadCloser) {
	dir := createDir(repo)
	f := createDockerfile(dir, dockerfile)
	defer f.Close()
	
	body, _ := ioutil.ReadAll(content)
	_, err := f.Write(body)
	if err != nil {
		log.Fatalln(err)
	}
}

func fetchDockerfile(line string) string {
	url := buildUrl(line)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode == 200 {
		repoName, fileName := splitLine(line, ",")
		writeDockerfile(repoName, fileName, resp.Body)
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
	lines, linesCount := readFile(fileName)

	for idx, line := range lines {
		status := fetchDockerfile(line)
		log.Printf("progress: %d/%d, code: %s, repo: %s \n", idx, linesCount, status, strings.Split(line, ",")[1])
	}
}
