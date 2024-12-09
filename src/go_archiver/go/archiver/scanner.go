package archiver

import (
	"os"
	"path/filepath"
	"sync"
)

// ScanResult represents the result of scanning a single file
type ScanResult struct {
	FileInfo FileInfo
	Error    error
}

// Scan starts the file scanning process
func (a *Archiver) Scan() (<-chan ScanResult, error) {
	out := make(chan ScanResult)
	
	// Validate source path
	_, err := os.Stat(a.config.SourcePath)
	if err != nil {
		return nil, err
	}
	
	go func() {
		defer close(out)
		
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 10) // Limit concurrent goroutines
		
		var scan func(string)
		scan = func(path string) {
			defer wg.Done()
			
			entries, err := os.ReadDir(path)
			if err != nil {
				out <- ScanResult{Error: err}
				return
			}
			
			for _, entry := range entries {
				entryPath := filepath.Join(path, entry.Name())
				info, err := entry.Info()
				if err != nil {
					out <- ScanResult{Error: err}
					continue
				}
				
				if info.IsDir() && a.config.Recursive {
					wg.Add(1)
					go func(p string) {
						semaphore <- struct{}{} // Acquire
						scan(p)
						<-semaphore // Release
					}(entryPath)
					continue
				}
				
				// Send file info through channel
				out <- ScanResult{
					FileInfo: FileInfo{
						Path:  entryPath,
						Size:  info.Size(),
						IsDir: info.IsDir(),
					},
				}
			}
		}
		
		wg.Add(1)
		scan(a.config.SourcePath)
		wg.Wait()
	}()
	
	return out, nil
}
