package archiver

import (
	"sync"
	"time"
)

type FilterMode string

const (
	FilterAll    FilterMode = "all"
	FilterPhotos FilterMode = "photos"
	FilterVideos FilterMode = "videos"
)

type FileType struct {
	Extension string
	MimeType  string
}

type Config struct {
	SourcePath  string
	OutputPath  string
	Recursive   bool
	FilterMode  FilterMode
	FileTypes   []string
	Modifiable  bool
}

type FileInfo struct {
	Path     string
	MimeType string
	Size     int64
	IsDir    bool
}

// Supported formats
var (
	PhotoFormats = map[string]FileType{
		"jpg":  {"jpg", "image/jpeg"},
		"jpeg": {"jpeg", "image/jpeg"},
		"png":  {"png", "image/png"},
		"webp": {"webp", "image/webp"},
		"heic": {"heic", "image/heic"},
		"heif": {"heif", "image/heif"},
	}

	VideoFormats = map[string]FileType{
		"mp4":  {"mp4", "video/mp4"},
		"webm": {"webm", "video/webm"},
		"avi":  {"avi", "video/x-msvideo"},
		"heif": {"heif", "video/heif"},
		"heic": {"heic", "video/heic"},
	}
)

// FileTypeCount tracks the number of files processed by type
type FileTypeCount struct {
	Photos map[string]int64 // maps extension to count
	Videos map[string]int64
	Others map[string]int64
}

// Result represents the processing result
type Result struct {
	StartTime      time.Time
	EndTime        time.Time
	FilesProcessed int64
	TotalSize     int64
	TypeCounts    FileTypeCount
	Progress      float64 // 0-100
	TotalFiles    int64  // For progress calculation
	Error         error
}

// Archiver handles the archiving process
type Archiver struct {
	config Config
	mu     sync.RWMutex
	result Result
}

// New creates a new Archiver instance
func New(config Config) *Archiver {
	return &Archiver{
		config: config,
		result: Result{
			StartTime: time.Now(),
			TypeCounts: FileTypeCount{
				Photos: make(map[string]int64),
				Videos: make(map[string]int64),
				Others: make(map[string]int64),
			},
		},
	}
}

// UpdateResult safely updates the result with new values
func (a *Archiver) UpdateResult(filesProcessed, totalSize int64, fileExt string, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if err != nil {
		a.result.Error = err
		return
	}
	
	a.result.FilesProcessed += filesProcessed
	a.result.TotalSize += totalSize
	
	// Update type counts
	if _, ok := PhotoFormats[fileExt]; ok {
		a.result.TypeCounts.Photos[fileExt]++
	} else if _, ok := VideoFormats[fileExt]; ok {
		a.result.TypeCounts.Videos[fileExt]++
	} else {
		a.result.TypeCounts.Others[fileExt]++
	}
	
	// Update progress
	if a.result.TotalFiles > 0 {
		a.result.Progress = float64(a.result.FilesProcessed) / float64(a.result.TotalFiles) * 100
	}
}

// SetTotalFiles sets the total number of files for progress calculation
func (a *Archiver) SetTotalFiles(total int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.result.TotalFiles = total
}

// GetProgress returns the current progress percentage
func (a *Archiver) GetProgress() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.result.Progress
}

// GetTypeCount returns the count of files by type
func (a *Archiver) GetTypeCount() FileTypeCount {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.result.TypeCounts
}

// GetDuration returns the elapsed processing time
func (a *Archiver) GetDuration() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.result.EndTime.IsZero() {
		return time.Since(a.result.StartTime)
	}
	return a.result.EndTime.Sub(a.result.StartTime)
}

// Finish marks the operation as complete
func (a *Archiver) Finish() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.result.EndTime = time.Now()
}
