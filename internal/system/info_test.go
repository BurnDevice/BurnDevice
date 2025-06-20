package system

import (
	"runtime"
	"testing"
)

func TestNewSystemInfo(t *testing.T) {
	sysInfo := NewSystemInfo()
	if sysInfo == nil {
		t.Fatal("Expected SystemInfo to be created")
	}
}

func TestCollect(t *testing.T) {
	sysInfo := NewSystemInfo()
	info, err := sysInfo.Collect()
	if err != nil {
		t.Fatalf("Failed to collect system info: %v", err)
	}

	if info == nil {
		t.Fatal("Expected info to be collected")
	}

	// Verify basic system information
	if info.OS != runtime.GOOS {
		t.Errorf("Expected OS %s, got %s", runtime.GOOS, info.OS)
	}

	if info.Architecture != runtime.GOARCH {
		t.Errorf("Expected architecture %s, got %s", runtime.GOARCH, info.Architecture)
	}

	if info.Hostname == "" {
		t.Error("Expected hostname to be set")
	}

	// Critical paths should be populated
	if len(info.CriticalPaths) == 0 {
		t.Error("Expected critical paths to be populated")
	}

	// Resources should be initialized
	if info.Resources.TotalMemory < 0 {
		t.Error("Expected total memory to be non-negative")
	}
}

func TestGetCriticalPaths(t *testing.T) {
	sysInfo := NewSystemInfo()
	paths := sysInfo.getCriticalPaths()

	if len(paths) == 0 {
		t.Error("Expected critical paths to be found")
	}

	// Check that paths are OS-appropriate
	switch runtime.GOOS {
	case "linux":
		found := false
		for _, path := range paths {
			if path == "/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected root path '/' to be in critical paths on Linux")
		}
	case "windows":
		found := false
		for _, path := range paths {
			if path == "C:\\Windows" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'C:\\Windows' to be in critical paths on Windows")
		}
	}
}

func TestGetRunningServices(t *testing.T) {
	sysInfo := NewSystemInfo()
	services, err := sysInfo.getRunningServices()

	// Services collection might fail on some systems, that's okay
	if err != nil {
		t.Logf("Service collection failed (expected on some systems): %v", err)
		return
	}

	if services == nil {
		t.Error("Expected services slice to be initialized")
	}

	t.Logf("Found %d running services", len(services))
}

func TestGetProcessList(t *testing.T) {
	sysInfo := NewSystemInfo()
	processes, err := sysInfo.getProcessList()

	// Process listing might fail on some systems
	if err != nil {
		t.Logf("Process listing failed (expected on some systems): %v", err)
		return
	}

	if processes == nil {
		t.Error("Expected processes slice to be initialized")
	}

	t.Logf("Found %d running processes", len(processes))
}

func TestGetResources(t *testing.T) {
	sysInfo := NewSystemInfo()
	resources, err := sysInfo.getResources()

	if err != nil {
		t.Logf("Resource collection failed: %v", err)
		return
	}

	if resources.TotalMemory < 0 {
		t.Error("Expected total memory to be non-negative")
	}

	if resources.AvailableMemory < 0 {
		t.Error("Expected available memory to be non-negative")
	}

	if resources.TotalDisk < 0 {
		t.Error("Expected total disk to be non-negative")
	}

	if resources.AvailableDisk < 0 {
		t.Error("Expected available disk to be non-negative")
	}

	if resources.CPUUsage < 0 || resources.CPUUsage > 100 {
		t.Errorf("Expected CPU usage to be between 0-100, got %.2f", resources.CPUUsage)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("Expected 'banana' to be found in slice")
	}

	if contains(slice, "grape") {
		t.Error("Expected 'grape' not to be found in slice")
	}

	if contains([]string{}, "anything") {
		t.Error("Expected empty slice to not contain anything")
	}
}
