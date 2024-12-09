package archiver

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Common errors
var (
	ErrInvalidRequest = errors.New("invalid modification request")
	ErrNotModifiable  = errors.New("tarball is not modifiable")
	ErrFileNotFound   = errors.New("file not found in archive")
)

// ModifyOperation represents the type of modification
type ModifyOperation int

const (
	OperationAdd ModifyOperation = iota
	OperationRemove
	OperationUpdate
)

// ModifyRequest represents a modification request
type ModifyRequest struct {
	Operation ModifyOperation
	Path      string        // Path of file to add/remove/update
	NewPath   string        // New path for updates (optional)
	FileInfo  FileInfo      // File info for additions
}

// ModifyResult represents the result of a modification
type ModifyResult struct {
	Operation ModifyOperation
	Path      string
	Success   bool
	Error     error
}

// FileEntry represents a file in the tarball
type FileEntry struct {
	Header  *tar.Header
	Offset  int64
	Size    int64
	ModTime time.Time
}

// TarballInfo stores information about files in the tarball
type TarballInfo struct {
	Files map[string]FileEntry
	mu    sync.RWMutex
}

// CompressionLevel represents gzip compression levels
type CompressionLevel int

const (
	CompressionDefault = CompressionLevel(gzip.DefaultCompression)
	CompressionBest    = CompressionLevel(gzip.BestCompression)
	CompressionFast    = CompressionLevel(gzip.BestSpeed)
	CompressionNone    = CompressionLevel(gzip.NoCompression)
)

// validateRequest validates a modification request
func (a *Archiver) validateRequest(req ModifyRequest) error {
	if !a.config.Modifiable {
		return ErrNotModifiable
	}

	switch req.Operation {
	case OperationAdd:
		if req.FileInfo.Path == "" {
			return errors.New("path required for add operation")
		}
		if _, err := os.Stat(req.FileInfo.Path); err != nil {
			return err
		}

	case OperationRemove:
		if req.Path == "" {
			return errors.New("path required for remove operation")
		}

	case OperationUpdate:
		if req.Path == "" || req.FileInfo.Path == "" {
			return errors.New("both old and new paths required for update operation")
		}
		if _, err := os.Stat(req.FileInfo.Path); err != nil {
			return err
		}

	default:
		return errors.New("invalid operation type")
	}

	return nil
}

// GetFileInfo retrieves information about a specific file in the tarball
func (a *Archiver) GetFileInfo(path string) (*FileEntry, error) {
	info, err := a.scanTarball()
	if err != nil {
		return nil, err
	}

	info.mu.RLock()
	defer info.mu.RUnlock()

	if entry, exists := info.Files[path]; exists {
		return &entry, nil
	}
	return nil, ErrFileNotFound
}

// ListFiles returns a list of all files in the tarball
func (a *Archiver) ListFiles() ([]string, error) {
	info, err := a.scanTarball()
	if err != nil {
		return nil, err
	}

	info.mu.RLock()
	defer info.mu.RUnlock()

	files := make([]string, 0, len(info.Files))
	for path := range info.Files {
		files = append(files, path)
	}
	return files, nil
}

// scanTarball scans the tarball and builds an index of files
func (a *Archiver) scanTarball() (*TarballInfo, error) {
	f, err := os.Open(a.config.OutputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	info := &TarballInfo{
		Files: make(map[string]FileEntry),
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		info.Files[header.Name] = FileEntry{
			Header:  header,
			Size:    header.Size,
			ModTime: header.ModTime,
		}
	}

	return info, nil
}

// Helper methods for file operations
func (a *Archiver) addFile(tw *tar.Writer, info FileInfo) error {
	file, err := os.Open(info.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    filepath.Base(info.Path),
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

func (a *Archiver) removeFile(tr *tar.Reader, tw *tar.Writer, path string) error {
	if tr == nil {
		return errors.New("tar reader is nil")
	}
	
	// Read through the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		// Skip the file we want to remove
		if header.Name == path {
			continue
		}
		
		// Copy other files to the new archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		
		if _, err := io.Copy(tw, tr); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *Archiver) updateFile(tr *tar.Reader, tw *tar.Writer, req ModifyRequest) error {
	updated := false
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Name == req.Path {
			// Replace with new file
			if err := a.addFile(tw, req.FileInfo); err != nil {
				return err
			}
			updated = true
			continue
		}

		// Copy other files
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(tw, tr); err != nil {
			return err
		}
	}

	if !updated {
		return ErrFileNotFound
	}
	return nil
}

// Modify modifies an existing tarball with concurrent processing
func (a *Archiver) Modify(requests []ModifyRequest, compression CompressionLevel) <-chan ModifyResult {
	out := make(chan ModifyResult)
	
	go func() {
		defer close(out)

		// Validate all requests first
		for _, req := range requests {
			if err := a.validateRequest(req); err != nil {
				out <- ModifyResult{
					Operation: req.Operation,
					Path:      req.Path,
					Error:     err,
				}
				return
			}
		}

		// Open original archive for reading
		srcFile, err := os.Open(a.config.OutputPath)
		if err != nil && !os.IsNotExist(err) {
			out <- ModifyResult{Error: err}
			return
		}
		defer srcFile.Close()

		// Create temporary file for the modified archive
		tempFile, err := os.CreateTemp(filepath.Dir(a.config.OutputPath), "temp_*.tar.gz")
		if err != nil {
			out <- ModifyResult{Error: err}
			return
		}
		tempPath := tempFile.Name()
		defer os.Remove(tempPath)

		// Set up writers
		gzw, err := gzip.NewWriterLevel(tempFile, int(compression))
		if err != nil {
			out <- ModifyResult{Error: err}
			return
		}
		defer gzw.Close()

		tw := tar.NewWriter(gzw)
		defer tw.Close()

		// Set up reader if source file exists
		var tr *tar.Reader
		if srcFile != nil {
			gzr, err := gzip.NewReader(srcFile)
			if err != nil {
				out <- ModifyResult{Error: err}
				return
			}
			defer gzr.Close()
			tr = tar.NewReader(gzr)
		}

		// Process modifications
		for _, req := range requests {
			switch req.Operation {
			case OperationAdd:
				if err := a.addFile(tw, req.FileInfo); err != nil {
					out <- ModifyResult{
						Operation: req.Operation,
						Path:      req.Path,
						Error:     err,
					}
					continue
				}
			case OperationRemove:
				if err := a.removeFile(tr, tw, req.Path); err != nil {
					out <- ModifyResult{
						Operation: req.Operation,
						Path:      req.Path,
						Error:     err,
					}
					continue
				}
			case OperationUpdate:
				if err := a.updateFile(tr, tw, req); err != nil {
					out <- ModifyResult{
						Operation: req.Operation,
						Path:      req.Path,
						Error:     err,
					}
					continue
				}
			}

			out <- ModifyResult{
				Operation: req.Operation,
				Path:      req.Path,
				Success:   true,
			}
		}

		// Replace original with modified version
		if err := os.Rename(tempPath, a.config.OutputPath); err != nil {
			out <- ModifyResult{Error: err}
			return
		}
	}()

	return out
}

// BulkModifyResult represents the result of a bulk modification operation
type BulkModifyResult struct {
	Successful int
	Failed     int
	Errors     []error
	Results    []ModifyResult
}

// BulkModify performs multiple modifications in batches
func (a *Archiver) BulkModify(requests []ModifyRequest, batchSize int, compression CompressionLevel) BulkModifyResult {
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	result := BulkModifyResult{
		Results: make([]ModifyResult, 0, len(requests)),
	}

	// Process requests in batches
	for i := 0; i < len(requests); i += batchSize {
		end := i + batchSize
		if end > len(requests) {
			end = len(requests)
		}

		// Process batch
		resultChan := a.Modify(requests[i:end], compression)
		for modResult := range resultChan {
			result.Results = append(result.Results, modResult)
			if modResult.Error != nil {
				result.Failed++
				result.Errors = append(result.Errors, modResult.Error)
			} else if modResult.Success {
				result.Successful++
			}
		}
	}

	return result
}

// BatchAddFiles adds multiple files to the archive in batches
func (a *Archiver) BatchAddFiles(paths []string, batchSize int, compression CompressionLevel) BulkModifyResult {
	requests := make([]ModifyRequest, len(paths))
	for i, path := range paths {
		requests[i] = ModifyRequest{
			Operation: OperationAdd,
			Path:      path,
			FileInfo: FileInfo{
				Path: path,
			},
		}
	}
	return a.BulkModify(requests, batchSize, compression)
}

// BatchRemoveFiles removes multiple files from the archive in batches
func (a *Archiver) BatchRemoveFiles(paths []string, batchSize int, compression CompressionLevel) BulkModifyResult {
	requests := make([]ModifyRequest, len(paths))
	for i, path := range paths {
		requests[i] = ModifyRequest{
			Operation: OperationRemove,
			Path:      path,
		}
	}
	return a.BulkModify(requests, batchSize, compression)
}
