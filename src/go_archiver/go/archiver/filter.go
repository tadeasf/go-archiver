package archiver

import (
	"path/filepath"
	"strings"
)

// FilterResult represents a filtered file
type FilterResult struct {
	FileInfo FileInfo
	Error    error
}

// Filter processes files based on configured filters
func (a *Archiver) Filter(results <-chan ScanResult) <-chan FilterResult {
	out := make(chan FilterResult)

	go func() {
		defer close(out)

		for result := range results {
			if result.Error != nil {
				out <- FilterResult{Error: result.Error}
				continue
			}

			// Get file extension
			ext := strings.ToLower(filepath.Ext(result.FileInfo.Path))
			if len(ext) > 0 {
				ext = ext[1:] // Remove the dot
			}

			// Check if file should be included based on filter mode
			include := false
			switch a.config.FilterMode {
			case FilterAll:
				include = true
				if len(a.config.FileTypes) > 0 {
					include = false
					for _, allowedType := range a.config.FileTypes {
						if ext == allowedType {
							include = true
							break
						}
					}
				}
			case FilterPhotos:
				include = isPhotoFile(ext)
			case FilterVideos:
				include = isVideoFile(ext)
			}

			if include {
				out <- FilterResult{
					FileInfo: result.FileInfo,
				}
			}
		}
	}()

	return out
}

func isPhotoFile(ext string) bool {
	photoExts := map[string]bool{
		"jpg": true, "jpeg": true, "png": true,
	}
	return photoExts[ext]
}

func isVideoFile(ext string) bool {
	videoExts := map[string]bool{
		"mp4": true, "avi": true, "mkv": true,
		"mov": true, "webm": true, "flv": true,
	}
	return videoExts[ext]
}
