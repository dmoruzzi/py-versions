package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

func NewVersion(versionStr string) Version {
	var major, minor, patch int
	fmt.Sscanf(versionStr, "%d.%d.%d", &major, &minor, &patch)
	return Version{Major: major, Minor: minor, Patch: patch}
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) LessThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}

type Versions []Version

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Versions) Less(i, j int) bool {
	return v[i].LessThan(v[j])
}

func fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractVersions(html string) map[string]map[string]interface{} {
	re := regexp.MustCompile(`href="(\d+\.\d+\.\d+)/"`) // e.g. href="3.9.0/"
	matches := re.FindAllStringSubmatch(html, -1)
	versions := make(map[string]map[string]interface{})

	for _, match := range matches {
		if len(match) == 2 {
			version := NewVersion(match[1])
			pyVersion := fmt.Sprintf("%d.%d", version.Major, version.Minor)

			if _, ok := versions[pyVersion]; !ok {
				versions[pyVersion] = make(map[string]interface{})
				versions[pyVersion]["latest"] = ""
				versions[pyVersion]["versions"] = []string{}
			}

			// Update latest version
			latest := versions[pyVersion]["latest"].(string)
			if latest == "" || compareVersions(match[1], latest) > 0 {
				versions[pyVersion]["latest"] = match[1]
			}

			// Update versions list
			versionsList := versions[pyVersion]["versions"].([]string)
			versionsList = append(versionsList, match[1])
			versions[pyVersion]["versions"] = versionsList
		}
	}

	return versions
}

func compareVersions(v1, v2 string) int {
	version1 := NewVersion(v1)
	version2 := NewVersion(v2)

	if version1.Major != version2.Major {
		return version1.Major - version2.Major
	}
	if version1.Minor != version2.Minor {
		return version1.Minor - version2.Minor
	}
	return version1.Patch - version2.Patch
}

func writeJSONFile(data interface{}, filename string) error {
	if filename == "" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "") // Disable indentation for console output
		if err := encoder.Encode(data); err != nil {
			return err
		}
	} else {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var url, jsonFilename string
	flag.StringVar(&url, "url", "https://www.python.org/ftp/python/", "Python FTP Mirror URL")
	flag.StringVar(&jsonFilename, "o", "", "Filename to save JSON output (optional; if not provided, output to console)")
	flag.Parse()

	html, err := fetchHTML(url)
	if err != nil {
		log.Fatalf("failed to fetch HTML: %v", err)
	}

	versions := extractVersions(html)
	if err := writeJSONFile(versions, jsonFilename); err != nil {
		log.Fatalf("failed to write JSON file: %v", err)
	}
}
