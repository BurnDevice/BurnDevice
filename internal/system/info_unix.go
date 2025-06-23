//go:build unix

package system

import (
	"fmt"
	"syscall"
)

// DiskInfo represents disk statistics
type DiskInfo struct {
	Total     int64
	Available int64
}

// getDiskInfo gets disk space information for Unix systems
func (s *SystemInfo) getDiskInfo() (*DiskInfo, error) {
	var stat syscall.Statfs_t
	path := "/"

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	// Safe conversion with bounds checking to prevent integer overflow
	// First check if values fit in int64 range
	const maxInt64 = int64(1<<63 - 1)

	if int64(stat.Blocks) > maxInt64 {
		return nil, fmt.Errorf("blocks value %d exceeds int64 maximum", stat.Blocks)
	}
	if int64(stat.Bsize) > maxInt64 {
		return nil, fmt.Errorf("block size value %d exceeds int64 maximum", stat.Bsize)
	}
	if int64(stat.Bavail) > maxInt64 {
		return nil, fmt.Errorf("available blocks value %d exceeds int64 maximum", stat.Bavail)
	}

	blocks := int64(stat.Blocks)
	bsize := int64(stat.Bsize)
	bavail := int64(stat.Bavail)

	// Check for potential overflow before multiplication
	if blocks > 0 && bsize > 0 && blocks > maxInt64/bsize {
		return nil, fmt.Errorf("disk size calculation would overflow")
	}
	if bavail > 0 && bsize > 0 && bavail > maxInt64/bsize {
		return nil, fmt.Errorf("available disk calculation would overflow")
	}

	total := blocks * bsize
	available := bavail * bsize

	return &DiskInfo{
		Total:     total,
		Available: available,
	}, nil
}
