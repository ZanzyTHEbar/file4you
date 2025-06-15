package trees

import (
	"time"
)

// FileMetadata represents metadata for a file in the filesystem
type FileMetadata struct {
	FilePath string    // Path to the file
	Size     int64     // Size in bytes
	ModTime  time.Time // Last modification time
	IsDir    bool      // Whether this is a directory
	Checksum string    // Optional checksum of file contents
}
