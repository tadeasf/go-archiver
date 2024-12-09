package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var testFiles = map[string][]byte{
    "images/photo1.jpg":     []byte("fake jpg content"),
    "images/photo2.png":     []byte("fake png content"),
    "images/photo3.webp":    []byte("fake webp content"),
    "videos/video1.mp4":     []byte("fake mp4 content"),
    "videos/video2.webm":    []byte("fake webm content"),
    "documents/doc1.txt":    []byte("This is a test document"),
    "documents/doc2.pdf":    []byte("fake pdf content"),
}

func main() {
    // Create base directories
    dirs := []string{
        "source",
        "source/images",
        "source/videos",
        "source/documents",
        "output",
    }

    for _, dir := range dirs {
        path := filepath.Join("testdata", dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            fmt.Printf("Error creating directory %s: %v\n", path, err)
            os.Exit(1)
        }
    }

    // Create test files
    for path, content := range testFiles {
        fullPath := filepath.Join("testdata/source", path)
        if err := os.WriteFile(fullPath, content, 0644); err != nil {
            fmt.Printf("Error creating file %s: %v\n", fullPath, err)
            os.Exit(1)
        }
    }

    // Create .gitkeep in output directory
    gitkeepPath := filepath.Join("testdata/output", ".gitkeep")
    if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
        fmt.Printf("Error creating .gitkeep: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Test files created successfully")
} 