package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"dagger/hugo/internal/dagger"
)

// PresentationStructure represents the YAML structure for presentation slides
type PresentationStructure struct {
	Slides map[string]Slide `json:"slides"`
}

// Slide represents a single slide in the presentation
type Slide struct {
	Order int    `json:"order"`
	File  string `json:"file"`
}

// MergeMarkdowns reads a presentation YAML file, fetches all referenced markdowns
// (either from src directory or via HTTP), sorts them by order, and returns a merged markdown file
func (m *Hugo) MergeMarkdowns(
	ctx context.Context,
	src *dagger.Directory,
	presentationFile *dagger.File) (*dagger.File, error) {
	// Read the presentation file
	presentationContent, err := presentationFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read presentation file: %w", err)
	}

	// Parse YAML manually since we don't have yaml library
	presentation, err := parseYAMLPresentation(presentationContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presentation YAML: %w", err)
	}

	// Create slice of slides with their content
	type slideWithContent struct {
		name    string
		order   int
		content string
	}

	var slides []slideWithContent

	// Process each slide
	for name, slide := range presentation.Slides {
		var content string

		// Check if file is URL or local path
		if strings.HasPrefix(slide.File, "http://") || strings.HasPrefix(slide.File, "https://") {
			// Download from URL
			resp, err := http.Get(slide.File)
			if err != nil {
				return nil, fmt.Errorf("failed to download %s: %w", slide.File, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("bad response from %s: %s", slide.File, resp.Status)
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body from %s: %w", slide.File, err)
			}

			content = string(bodyBytes)
			fmt.Printf("✅ Downloaded: %s\n", slide.File)
		} else {
			// Read from src directory
			fileContent, err := src.File(slide.File).Contents(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s from src: %w", slide.File, err)
			}

			content = fileContent
			fmt.Printf("✅ Read: %s\n", slide.File)
		}

		slides = append(slides, slideWithContent{
			name:    name,
			order:   slide.Order,
			content: content,
		})
	}

	// Sort slides by order
	sort.Slice(slides, func(i, j int) bool {
		return slides[i].order < slides[j].order
	})

	// Merge all contents
	var merged bytes.Buffer
	for i, slide := range slides {
		if i > 0 {
			merged.WriteString("\n\n---\n\n")
		}
		merged.WriteString(slide.content)
	}

	// Create output file using container
	mergedFile := dag.Container().
		From("alpine:latest").
		WithNewFile("/merged.md", merged.String()).
		WithoutEntrypoint().
		File("/merged.md")

	return mergedFile, nil
}

// parseYAMLPresentation parses the YAML presentation structure manually
func parseYAMLPresentation(yamlContent string) (*PresentationStructure, error) {
	presentation := &PresentationStructure{
		Slides: make(map[string]Slide),
	}

	lines := strings.Split(yamlContent, "\n")
	var currentSlide string
	var currentOrder int
	var currentFile string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip "slides:" line
		if line == "slides:" {
			continue
		}

		// Check for slide name (starts with no indent after removing spaces)
		if !strings.HasPrefix(line, "order:") && !strings.HasPrefix(line, "file:") && strings.Contains(line, ":") && !strings.HasPrefix(line, " ") {
			// Save previous slide if any
			if currentSlide != "" {
				presentation.Slides[currentSlide] = Slide{Order: currentOrder, File: currentFile}
			}
			// New slide
			currentSlide = strings.TrimSuffix(line, ":")
			currentOrder = 0
			currentFile = ""
			continue
		}

		// Parse order field
		if strings.HasPrefix(line, "order:") {
			orderStr := strings.TrimPrefix(line, "order:")
			orderStr = strings.TrimSpace(orderStr)
			_, _ = fmt.Sscanf(orderStr, "%d", &currentOrder)
			continue
		}

		// Parse file field
		if strings.HasPrefix(line, "file:") {
			currentFile = strings.TrimPrefix(line, "file:")
			currentFile = strings.TrimSpace(currentFile)
			continue
		}
	}

	// Save last slide
	if currentSlide != "" {
		presentation.Slides[currentSlide] = Slide{Order: currentOrder, File: currentFile}
	}

	return presentation, nil
}
