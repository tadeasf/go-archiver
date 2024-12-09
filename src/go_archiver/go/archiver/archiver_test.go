package archiver

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
    config := Config{
        SourcePath:  "testdata/source",
        OutputPath:  "testdata/output.tar.gz",
        Recursive:   true,
        FilterMode:  FilterAll,
        Modifiable:  true,
        FileTypes:   []string{"jpg", "png"},
    }

    a := New(config)
    if a == nil {
        t.Fatal("Expected non-nil Archiver")
    }

    // Compare fields individually since Config contains a slice
    if a.config.SourcePath != config.SourcePath {
        t.Errorf("Expected SourcePath %v, got %v", config.SourcePath, a.config.SourcePath)
    }
    if a.config.OutputPath != config.OutputPath {
        t.Errorf("Expected OutputPath %v, got %v", config.OutputPath, a.config.OutputPath)
    }
    if a.config.Recursive != config.Recursive {
        t.Errorf("Expected Recursive %v, got %v", config.Recursive, a.config.Recursive)
    }
    if a.config.FilterMode != config.FilterMode {
        t.Errorf("Expected FilterMode %v, got %v", config.FilterMode, a.config.FilterMode)
    }
    if a.config.Modifiable != config.Modifiable {
        t.Errorf("Expected Modifiable %v, got %v", config.Modifiable, a.config.Modifiable)
    }
    if !reflect.DeepEqual(a.config.FileTypes, config.FileTypes) {
        t.Errorf("Expected FileTypes %v, got %v", config.FileTypes, a.config.FileTypes)
    }
}

// Common test cases to ensure consistency across all tests
var commonTestCases = []struct {
    name          string
    mode          FilterMode
    fileTypes     []string
    expected      int
    expectedFiles []string
}{
    {
        name:     "all files",
        mode:     FilterAll,
        expected: 7,  // All files in testdata/source
    },
    {
        name:     "photos only",
        mode:     FilterPhotos,
        expected: 2,  // Only jpg and png files
        expectedFiles: []string{
            "images/photo1.jpg",
            "images/photo2.png",
        },
    },
    {
        name:     "videos only",
        mode:     FilterVideos,
        expected: 2,  // mp4 and webm files
        expectedFiles: []string{
            "videos/video1.mp4",
            "videos/video2.webm",
        },
    },
    {
        name:      "specific types",
        mode:      FilterAll,
        fileTypes: []string{"jpg"},
        expected:  1,  // Only jpg files
        expectedFiles: []string{
            "images/photo1.jpg",
        },
    },
}

func TestScan(t *testing.T) {
    sourceDir := "testdata/source"
    outputDir := "testdata/output"

    for _, tt := range commonTestCases {
        t.Run(tt.name, func(t *testing.T) {
            config := Config{
                SourcePath:  sourceDir,
                OutputPath:  filepath.Join(outputDir, "test.tar.gz"),
                Recursive:   true,
                FilterMode: tt.mode,
                FileTypes:  tt.fileTypes,
            }

            a := New(config)
            scanResults, err := a.Scan()
            if err != nil {
                t.Fatal(err)
            }

            filterResults := a.Filter(scanResults)
            count := 0
            foundFiles := make(map[string]bool)
            
            for result := range filterResults {
                if result.Error != nil {
                    t.Errorf("Unexpected error: %v", result.Error)
                }
                relPath, err := filepath.Rel(sourceDir, result.FileInfo.Path)
                if err != nil {
                    t.Fatal(err)
                }
                count++
                foundFiles[relPath] = true
            }

            if count != tt.expected {
                t.Errorf("Expected %d files, got %d", tt.expected, count)
            }

            if tt.expectedFiles != nil {
                for _, expectedFile := range tt.expectedFiles {
                    if !foundFiles[expectedFile] {
                        t.Errorf("Expected file %s not found in filtered results", expectedFile)
                    }
                }
            }
        })
    }
}

func TestModify(t *testing.T) {
    sourceDir := "testdata/source"
    outputDir := "testdata/output"
    outputPath := filepath.Join(outputDir, "test.tar.gz")

    for _, tt := range commonTestCases {
        t.Run(tt.name, func(t *testing.T) {
            config := Config{
                SourcePath:  sourceDir,
                OutputPath:  outputPath,
                Recursive:   true,
                FilterMode: tt.mode,
                FileTypes:  tt.fileTypes,
                Modifiable: true,
            }

            a := New(config)
            scanResults, err := a.Scan()
            if err != nil {
                t.Fatal(err)
            }

            filterResults := a.Filter(scanResults)
            count := 0
            foundFiles := make(map[string]bool)
            
            for result := range filterResults {
                if result.Error != nil {
                    t.Errorf("Unexpected error: %v", result.Error)
                }
                relPath, err := filepath.Rel(sourceDir, result.FileInfo.Path)
                if err != nil {
                    t.Fatal(err)
                }
                count++
                foundFiles[relPath] = true
            }

            if count != tt.expected {
                t.Errorf("Expected %d files, got %d", tt.expected, count)
            }

            if tt.expectedFiles != nil {
                for _, expectedFile := range tt.expectedFiles {
                    if !foundFiles[expectedFile] {
                        t.Errorf("Expected file %s not found in filtered results", expectedFile)
                    }
                }
            }
        })
    }
}

func TestBulkModify(t *testing.T) {
    sourceDir := "testdata/source"
    outputDir := "testdata/output"
    outputPath := filepath.Join(outputDir, "test.tar.gz")

    for _, tt := range commonTestCases {
        t.Run(tt.name, func(t *testing.T) {
            config := Config{
                SourcePath:  sourceDir,
                OutputPath:  outputPath,
                Recursive:   true,
                FilterMode: tt.mode,
                FileTypes:  tt.fileTypes,
                Modifiable: true,
            }

            a := New(config)
            scanResults, err := a.Scan()
            if err != nil {
                t.Fatal(err)
            }

            filterResults := a.Filter(scanResults)
            count := 0
            foundFiles := make(map[string]bool)
            
            for result := range filterResults {
                if result.Error != nil {
                    t.Errorf("Unexpected error: %v", result.Error)
                }
                relPath, err := filepath.Rel(sourceDir, result.FileInfo.Path)
                if err != nil {
                    t.Fatal(err)
                }
                count++
                foundFiles[relPath] = true
            }

            if count != tt.expected {
                t.Errorf("Expected %d files, got %d", tt.expected, count)
            }

            if tt.expectedFiles != nil {
                for _, expectedFile := range tt.expectedFiles {
                    if !foundFiles[expectedFile] {
                        t.Errorf("Expected file %s not found in filtered results", expectedFile)
                    }
                }
            }
        })
    }
} 