package archiver

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// CreateResult represents the result of archive creation
type CreateResult struct {
	FilesProcessed int64
	TotalSize     int64
	Error         error
}

// Create generates a tarball from the filtered files
func (a *Archiver) Create(in <-chan FilterResult) <-chan CreateResult {
	out := make(chan CreateResult)

	go func() {
		defer close(out)

		// Create the output file
		f, err := os.Create(a.config.OutputPath)
		if err != nil {
			out <- CreateResult{Error: err}
			return
		}
		defer f.Close()

		// Create gzip writer
		gw := gzip.NewWriter(f)
		defer gw.Close()

		// Create tar writer
		tw := tar.NewWriter(gw)
		defer tw.Close()

		var (
			filesProcessed int64
			totalSize     int64
			wg           sync.WaitGroup
			errChan      = make(chan error, 1)
			semaphore    = make(chan struct{}, 5) // Limit concurrent file processing
		)

		// Process files concurrently
		for result := range in {
			if result.Error != nil {
				out <- CreateResult{Error: result.Error}
				continue
			}

			wg.Add(1)
			go func(res FilterResult) {
				defer wg.Done()
				semaphore <- struct{}{} // Acquire
				defer func() { <-semaphore }() // Release

				// Open and process the file
				if err := a.addFileToTar(tw, res.FileInfo); err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}

				atomic.AddInt64(&filesProcessed, 1)
				atomic.AddInt64(&totalSize, res.FileInfo.Size)
			}(result)
		}

		// Wait for all files to be processed
		wg.Wait()

		// Check for any errors
		select {
		case err := <-errChan:
			out <- CreateResult{Error: err}
			return
		default:
			out <- CreateResult{
				FilesProcessed: filesProcessed,
				TotalSize:     totalSize,
			}
		}
	}()

	return out
}

// addFileToTar adds a single file to the tar archive
func (a *Archiver) addFileToTar(tw *tar.Writer, info FileInfo) error {
	file, err := os.Open(info.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create tar header
	header := &tar.Header{
		Name:    filepath.Base(info.Path),
		Size:    info.Size,
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// Copy file content to tar
	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}
