package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

// Extract all <img src="..."> links from a markdown string
func extractImageLinks(markdown string) []string {
	re := regexp.MustCompile(`<img\s+[^>]*src=["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(markdown, -1)

	var links []string
	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, match[1])
		}
	}
	return links
}

// Download each image from the list and save it to destDir
func downloadImages(urls []string, destDir string) error {
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad response from %s: %s", url, resp.Status)
		}

		filename := path.Base(url)
		outPath := filepath.Join(destDir, filename)

		outFile, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outPath, err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to save file %s: %w", outPath, err)
		}

		fmt.Printf("✅ Downloaded: %s\n", outPath)
	}
	return nil
}

// Replace all <img src="..."> URLs with local paths from imageMap
func replaceImageLinks(markdown string, imageMap map[string]string) string {
	re := regexp.MustCompile(`<img\s+([^>]*?)src=["']([^"']+)["']`)
	return re.ReplaceAllStringFunc(markdown, func(imgTag string) string {
		matches := re.FindStringSubmatch(imgTag)
		if len(matches) == 3 {
			attrs := matches[1]
			originalURL := matches[2]
			if localPath, ok := imageMap[originalURL]; ok {
				return fmt.Sprintf(`<img %ssrc="%s"`, attrs, localPath)
			}
		}
		return imgTag // fallback to original
	})
}

func downloadImageAsBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return bytes, nil
}
