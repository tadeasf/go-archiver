package bindings

import (
	"go-archiver/archiver"
)

// PyArchiver wraps the Go Archiver for Python
type PyArchiver struct {
    arch *archiver.Archiver
}

// NewArchiver creates a new PyArchiver instance
func NewArchiver(sourcePath, outputPath string, recursive bool, filterMode string) *PyArchiver {
    config := archiver.Config{
        SourcePath:  sourcePath,
        OutputPath:  outputPath,
        Recursive:   recursive,
        FilterMode:  archiver.FilterMode(filterMode),
        Modifiable:  true,
    }
    
    return &PyArchiver{
        arch: archiver.New(config),
    }
}

// Archive processes files and creates the archive
func (p *PyArchiver) Archive() error {
    scanResults, err := p.arch.Scan()
    if err != nil {
        return err
    }

    filterResults := p.arch.Filter(scanResults)
    createResults := p.arch.Create(filterResults)

    // Wait for completion
    for result := range createResults {
        if result.Error != nil {
            return result.Error
        }
    }

    return nil
}