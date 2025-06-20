//go:build windows

package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// DiskInfo represents disk statistics
type DiskInfo struct {
	Total     int64
	Available int64
}

// getDiskInfo gets disk space information for Windows systems
func (s *SystemInfo) getDiskInfo() (*DiskInfo, error) {
	// Use wmic to get disk space information
	cmd := exec.Command("wmic", "logicaldisk", "where", "caption=\"C:\"", "get", "size,freespace", "/value")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %v", err)
	}

	var total, available int64
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FreeSpace=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				available, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse free space: %v", err)
				}
			}
		} else if strings.HasPrefix(line, "Size=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				total, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse total size: %v", err)
				}
			}
		}
	}

	if total == 0 {
		return nil, fmt.Errorf("failed to get disk size information")
	}

	return &DiskInfo{
		Total:     total,
		Available: available,
	}, nil
}
