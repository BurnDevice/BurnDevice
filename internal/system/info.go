package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SystemInfo collects system information
type SystemInfo struct{}

// Info represents collected system information
type Info struct {
	OS              string
	Architecture    string
	Hostname        string
	CriticalPaths   []string
	RunningServices []string
	Resources       Resources
}

// Resources represents system resource information
type Resources struct {
	TotalMemory     int64
	AvailableMemory int64
	TotalDisk       int64
	AvailableDisk   int64
	CPUUsage        float64
}

// NewSystemInfo creates a new system info collector
func NewSystemInfo() *SystemInfo {
	return &SystemInfo{}
}

// Collect gathers comprehensive system information
func (s *SystemInfo) Collect() (*Info, error) {
	info := &Info{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	info.Hostname = hostname

	// Collect critical paths
	info.CriticalPaths = s.getCriticalPaths()

	// Collect running services
	services, err := s.getRunningServices()
	if err == nil {
		info.RunningServices = services
	}

	// Collect resource information
	resources, err := s.getResources()
	if err == nil {
		info.Resources = resources
	}

	return info, nil
}

// getCriticalPaths returns a list of critical system paths
func (s *SystemInfo) getCriticalPaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "linux":
		paths = []string{
			"/",
			"/boot",
			"/bin",
			"/sbin",
			"/usr",
			"/etc",
			"/var",
			"/proc",
			"/sys",
			"/dev",
		}
	case "windows":
		paths = []string{
			"C:\\Windows",
			"C:\\Windows\\System32",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
			"C:\\Users",
		}
	case "darwin":
		paths = []string{
			"/",
			"/System",
			"/Library",
			"/usr",
			"/bin",
			"/sbin",
			"/Applications",
		}
	}

	// Filter existing paths
	var existingPaths []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	return existingPaths
}

// getRunningServices returns a list of running services
func (s *SystemInfo) getRunningServices() ([]string, error) {
	var services []string

	switch runtime.GOOS {
	case "linux":
		// Use systemctl to list services
		cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=running", "--no-legend")
		output, err := cmd.Output()
		if err != nil {
			// Fallback to ps command
			return s.getProcessList()
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) > 0 {
				serviceName := strings.TrimSuffix(fields[0], ".service")
				services = append(services, serviceName)
			}
		}

	case "windows":
		// Use sc query to list services
		cmd := exec.Command("sc", "query", "type=", "service", "state=", "running")
		output, err := cmd.Output()
		if err != nil {
			return s.getProcessList()
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "SERVICE_NAME:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					serviceName := strings.TrimSpace(parts[1])
					services = append(services, serviceName)
				}
			}
		}

	default:
		return s.getProcessList()
	}

	return services, nil
}

// getProcessList returns a list of running processes as fallback
func (s *SystemInfo) getProcessList() ([]string, error) {
	var processes []string

	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}
		fields := strings.Fields(line)
		if len(fields) > 10 {
			processName := fields[10]
			// Only include unique process names
			if !contains(processes, processName) {
				processes = append(processes, processName)
			}
		}
	}

	return processes, nil
}

// getResources collects system resource information
func (s *SystemInfo) getResources() (Resources, error) {
	resources := Resources{}

	// Get memory information
	memInfo, err := s.getMemoryInfo()
	if err == nil {
		resources.TotalMemory = memInfo.Total
		resources.AvailableMemory = memInfo.Available
	}

	// Get disk information
	diskInfo, err := s.getDiskInfo()
	if err == nil {
		resources.TotalDisk = diskInfo.Total
		resources.AvailableDisk = diskInfo.Available
	}

	// Get CPU usage
	cpuUsage, err := s.getCPUUsage()
	if err == nil {
		resources.CPUUsage = cpuUsage
	}

	return resources, nil
}

// MemoryInfo represents memory statistics
type MemoryInfo struct {
	Total     int64
	Available int64
}

// getMemoryInfo collects memory information
func (s *SystemInfo) getMemoryInfo() (*MemoryInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return s.getLinuxMemoryInfo()
	case "windows":
		return s.getWindowsMemoryInfo()
	case "darwin":
		return s.getDarwinMemoryInfo()
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// getLinuxMemoryInfo reads memory info from /proc/meminfo
func (s *SystemInfo) getLinuxMemoryInfo() (*MemoryInfo, error) {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	memInfo := &MemoryInfo{}

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				total, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					memInfo.Total = total * 1024 // Convert KB to bytes
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				available, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					memInfo.Available = available * 1024 // Convert KB to bytes
				}
			}
		}
	}

	return memInfo, nil
}

// getWindowsMemoryInfo gets Windows memory information
func (s *SystemInfo) getWindowsMemoryInfo() (*MemoryInfo, error) {
	// This is a simplified implementation
	// In production, you would use Windows API calls
	cmd := exec.Command("wmic", "computersystem", "get", "TotalPhysicalMemory", "/value")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "TotalPhysicalMemory=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				total, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err == nil {
					return &MemoryInfo{
						Total:     total,
						Available: total / 2, // Rough estimate
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("failed to parse memory info")
}

// getDarwinMemoryInfo gets macOS memory information
func (s *SystemInfo) getDarwinMemoryInfo() (*MemoryInfo, error) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	total, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return nil, err
	}

	return &MemoryInfo{
		Total:     total,
		Available: total / 2, // Rough estimate
	}, nil
}

// getCPUUsage gets current CPU usage percentage
func (s *SystemInfo) getCPUUsage() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return s.getLinuxCPUUsage()
	case "windows":
		return s.getWindowsCPUUsage()
	case "darwin":
		return s.getDarwinCPUUsage()
	default:
		return 0.0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// getLinuxCPUUsage gets CPU usage on Linux
func (s *SystemInfo) getLinuxCPUUsage() (float64, error) {
	cmd := exec.Command("grep", "^cpu", "/proc/stat")
	output, err := cmd.Output()
	if err != nil {
		return 0.0, err
	}

	line := strings.TrimSpace(string(output))
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return 0.0, fmt.Errorf("invalid /proc/stat format")
	}

	user, _ := strconv.ParseFloat(fields[1], 64)
	nice, _ := strconv.ParseFloat(fields[2], 64)
	system, _ := strconv.ParseFloat(fields[3], 64)
	idle, _ := strconv.ParseFloat(fields[4], 64)

	total := user + nice + system + idle
	if total == 0 {
		return 0.0, nil
	}

	return ((user + nice + system) / total) * 100, nil
}

// getWindowsCPUUsage gets CPU usage on Windows
func (s *SystemInfo) getWindowsCPUUsage() (float64, error) {
	cmd := exec.Command("wmic", "cpu", "get", "loadpercentage", "/value")
	output, err := cmd.Output()
	if err != nil {
		return 0.0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LoadPercentage=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				usage, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err == nil {
					return usage, nil
				}
			}
		}
	}

	return 0.0, fmt.Errorf("failed to parse CPU usage")
}

// getDarwinCPUUsage gets CPU usage on macOS
func (s *SystemInfo) getDarwinCPUUsage() (float64, error) {
	cmd := exec.Command("top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		return 0.0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			// Parse CPU usage line
			parts := strings.Split(line, ",")
			for _, part := range parts {
				if strings.Contains(part, "% idle") {
					idleStr := strings.TrimSpace(strings.Replace(part, "% idle", "", 1))
					idle, err := strconv.ParseFloat(idleStr, 64)
					if err == nil {
						return 100.0 - idle, nil
					}
				}
			}
		}
	}

	return 0.0, fmt.Errorf("failed to parse CPU usage")
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
