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
	// Try wmic first
	diskInfo, err := s.getDiskInfoWmic()
	if err == nil {
		return diskInfo, nil
	}

	// Fallback to PowerShell
	return s.getDiskInfoPowerShell()
}

// getDiskInfoWmic uses wmic to get disk information
func (s *SystemInfo) getDiskInfoWmic() (*DiskInfo, error) {
	// Use wmic to get disk space information for C: drive
	cmd := exec.Command("wmic", "logicaldisk", "where", "caption=\"C:\"", "get", "size,freespace", "/format:list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info via wmic: %v", err)
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

// getDiskInfoPowerShell uses PowerShell to get disk information
func (s *SystemInfo) getDiskInfoPowerShell() (*DiskInfo, error) {
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_LogicalDisk -Filter \"DeviceID='C:'\" | Select-Object Size,FreeSpace | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info via PowerShell: %v", err)
	}

	// Simple JSON parsing (avoiding external dependencies)
	content := string(output)
	var total, available int64

	// Extract Size
	if strings.Contains(content, "Size") {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Size") && strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					sizeStr := strings.Trim(strings.TrimSpace(parts[1]), ",")
					if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
						total = size
					}
				}
			}
			if strings.Contains(line, "FreeSpace") && strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					freeStr := strings.Trim(strings.TrimSpace(parts[1]), ",")
					if free, err := strconv.ParseInt(freeStr, 10, 64); err == nil {
						available = free
					}
				}
			}
		}
	}

	if total == 0 {
		return nil, fmt.Errorf("failed to parse disk information from PowerShell")
	}

	return &DiskInfo{
		Total:     total,
		Available: available,
	}, nil
}
