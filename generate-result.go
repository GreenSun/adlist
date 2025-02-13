package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	includeDir = "include"
	excludeDir = "exclude"
	fileExt    = ".txt"
	resultFile = "result.txt"
	timeout    = 10 * time.Minute
	retryCount = 3
)

var urls = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://www.github.developerdan.com/hosts/lists/ads-and-tracking-extended.txt",
	"https://v.firebog.net/hosts/AdguardDNS.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt",
	"https://winhelp2002.mvps.org/hosts.txt",
	"https://sysctl.org/cameleon/hosts",
}

var (
	commentRegex      = regexp.MustCompile(`\s*#.*$`)
	multiSpacesRegexp = regexp.MustCompile(`\s\s+`)
	ipAddressRegex    = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+\s`)
	ipAddressHost     = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+\s\d+\.\d+\.\d+\.\d+`)
)

func main() {
	tempDir, err := os.MkdirTemp("", "adlist")
	if err != nil {
		log.Println("can not create temp dir:", err)
		return
	}
	defer os.RemoveAll(tempDir)

	os.MkdirAll(includeDir, os.ModePerm)
	os.MkdirAll(excludeDir, os.ModePerm)

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			filename := generateFilename(url)
			tempFile := filepath.Join(tempDir, filename)

			if downloadFile(url, tempFile) {
				if isFileNonEmpty(tempFile) {
					newFileName := filepath.Join(includeDir, filename)
					err = moveFile(tempFile, newFileName)
					if err != nil {
						log.Printf("renaming downloaded file error: %s from %s to %s %s\n", url, tempFile, newFileName, err)
					}
				} else {
					log.Printf("file is empty %s for URL: %s\n", tempFile, url)
				}
			} else {
				log.Printf("downloading error: %s\n", url)
			}
		}(url)
	}
	wg.Wait()

	includeLines := readFilesInDir(includeDir)
	excludeLines := readFilesInDir(excludeDir)

	filteredLines := filterLines(includeLines, excludeLines)

	saveResult(filteredLines, resultFile)
}

func generateFilename(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")
	url = strings.TrimSuffix(url, "/")
	url = regexp.MustCompile(`[^a-zA-Z0-9._-]+`).ReplaceAllString(url, "-")
	if !strings.HasSuffix(url, ".txt") {
		url += ".txt"
	}
	return url
}

func downloadFile(url, filePath string) bool {
	client := &http.Client{Timeout: timeout}
	for i := 0; i < retryCount; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err == nil {
				err = os.WriteFile(filePath, data, 0644)
				if err == nil {
					return true
				}
			}
		}
	}
	return false
}

func isFileNonEmpty(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && info.Size() > 0
}

func readFilesInDir(dir string) []string {
	var lines []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), fileExt) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := cleanLine(scanner.Text())
			if line != "" {
				lines = append(lines, line)
			}
		}
		return nil
	})
	if err != nil {
		log.Println("file reading error:", dir, err)
	}
	sort.Strings(lines)
	return unique(lines)
}

func cleanLine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") || strings.Contains(line, "::") {
		return ""
	}

	line = commentRegex.ReplaceAllString(line, "")

	if line == "" {
		return ""
	}

	line = multiSpacesRegexp.ReplaceAllString(line, " ")
	line = strings.Replace(line, "127.0.0.1 ", "0.0.0.0 ", 1)

	hasIPAddress := ipAddressRegex.MatchString(line)
	if !hasIPAddress {
		line = "0.0.0.0 " + line
	}

	hasIpAddressHost := ipAddressHost.MatchString(line)
	if hasIpAddressHost {
		return ""
	}

	return line
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func filterLines(include []string, exclude []string) []string {
	excludeMap := make(map[string]bool)
	for _, line := range exclude {
		excludeMap[line] = true
	}
	var filtered []string
	for _, line := range include {
		if !excludeMap[line] {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

func moveFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", source, err)
	}

	err = os.WriteFile(destination, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing file %s: %w", destination, err)
	}

	err = os.Remove(source)
	if err != nil {
		return fmt.Errorf("error removing file %s: %w", source, err)
	}

	return nil
}

func saveResult(lines []string, filePath string) {
	content := strings.Join(lines, "\n") + "\n"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Println("saving error:", err)
	}
}
