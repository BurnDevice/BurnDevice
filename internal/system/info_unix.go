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
	// Check if uint64 values fit in int64 range before conversion
	const maxInt64 = uint64(1<<63 - 1)

	if stat.Blocks > maxInt64 {
		return nil, fmt.Errorf("blocks value %d exceeds int64 maximum", stat.Blocks)
	}
	if stat.Bavail > maxInt64 {
		return nil, fmt.Errorf("available blocks value %d exceeds int64 maximum", stat.Bavail)
	}
	// Bsize is already int64, just check for negative values
	if stat.Bsize < 0 {
		return nil, fmt.Errorf("block size cannot be negative: %d", stat.Bsize)
	}

	// #nosec G115 - Safe conversion: bounds checked above
	blocks := int64(stat.Blocks)
	bsize := stat.Bsize // Already int64
	// #nosec G115 - Safe conversion: bounds checked above
	bavail := int64(stat.Bavail)

	// Check for potential overflow before multiplication
	if blocks > 0 && bsize > 0 && blocks > int64(maxInt64)/bsize {
		return nil, fmt.Errorf("disk size calculation would overflow")
	}
	if bavail > 0 && bsize > 0 && bavail > int64(maxInt64)/bsize {
		return nil, fmt.Errorf("available disk calculation would overflow")
	}

	total := blocks * bsize
	available := bavail * bsize

	return &DiskInfo{
		Total:     total,
		Available: available,
	}, nil
}
