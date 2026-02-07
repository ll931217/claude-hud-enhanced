package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/errors"
)

// ThresholdLevel represents a color threshold level
type ThresholdLevel int

const (
	LevelGood  ThresholdLevel = iota // Green (0-70%)
	LevelWarning                       // Yellow (70-90%)
	LevelCritical                      // Red (>90%)
)

// Monitor tracks system resources
type Monitor struct {
	mu            sync.RWMutex
	lastUpdate    time.Time
	updateInterval time.Duration
	cpu           CPUInfo
	memory        MemoryInfo
	disk          DiskInfo
	fd            FDInfo
	currentDir    string
	language      string
}

// CPUInfo contains CPU usage information
type CPUInfo struct {
	UsagePercent float64
	CoreCount    int
}

// MemoryInfo contains memory usage information
type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Percent   float64
}

// DiskInfo contains disk usage information
type DiskInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Percent   float64
	Path      string
}

// FDInfo contains file descriptor information
type FDInfo struct {
	Count int
}

// NewMonitor creates a new system monitor
func NewMonitor() *Monitor {
	return &Monitor{
		updateInterval: 5 * time.Second,
	}
}

// Update refreshes all system metrics
func (m *Monitor) Update() error {
	return errors.SafeCall(func() error {
		m.mu.Lock()
		defer m.mu.Unlock()

		// Check if we need to update (cache for 5 seconds)
		if time.Since(m.lastUpdate) < m.updateInterval && m.lastUpdate.IsZero() == false {
			return nil
		}

		// Update CPU
		if cpu, err := getCPUUsage(); err == nil {
			m.cpu = cpu
		}

		// Update Memory
		if mem, err := getMemoryUsage(); err == nil {
			m.memory = mem
		}

		// Update Disk
		if disk, err := getDiskUsage(); err == nil {
			m.disk = disk
		}

		// Update File Descriptors
		if fd, err := getFDCount(); err == nil {
			m.fd = fd
		}

		// Update current directory
		if cwd, err := os.Getwd(); err == nil {
			m.currentDir = cwd
		}

		// Update language detection
		m.language = detectLanguage(m.currentDir)

		m.lastUpdate = time.Now()
		return nil
	})
}

// GetCPU returns the current CPU usage
func (m *Monitor) GetCPU() CPUInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cpu
}

// GetMemory returns the current memory usage
func (m *Monitor) GetMemory() MemoryInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memory
}

// GetDisk returns the current disk usage
func (m *Monitor) GetDisk() DiskInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.disk
}

// GetFD returns the current file descriptor count
func (m *Monitor) GetFD() FDInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fd
}

// GetCurrentDir returns the current working directory
func (m *Monitor) GetCurrentDir() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentDir
}

// GetLanguage returns the detected programming language
func (m *Monitor) GetLanguage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.language
}

// GetThresholdLevel returns the color threshold level for a percentage
func GetThresholdLevel(percent float64) ThresholdLevel {
	if percent >= 90 {
		return LevelCritical
	} else if percent >= 70 {
		return LevelWarning
	}
	return LevelGood
}

// FormatCPUDisplay formats CPU usage for display
func (m *Monitor) FormatCPUDisplay() string {
	if m.cpu.UsagePercent == 0 {
		return ""
	}
	return fmt.Sprintf("CPU %.0f%%", m.cpu.UsagePercent)
}

// FormatMemoryDisplay formats memory usage for display
func (m *Monitor) FormatMemoryDisplay() string {
	if m.memory.Total == 0 {
		return ""
	}

	return fmt.Sprintf("RAM %.0f%%", m.memory.Percent)
}

// FormatDiskDisplay formats disk usage for display
func (m *Monitor) FormatDiskDisplay() string {
	if m.disk.Total == 0 {
		return ""
	}

	return fmt.Sprintf("DISK %.0f%%", m.disk.Percent)
}

// FormatFDDisplay formats file descriptor count for display
func (m *Monitor) FormatFDDisplay() string {
	if m.fd.Count == 0 {
		return ""
	}

	return fmt.Sprintf("FD %d", m.fd.Count)
}

// FormatDirDisplay formats the current directory for display
func (m *Monitor) FormatDirDisplay() string {
	if m.currentDir == "" {
		return ""
	}

	// Get the base directory name
	dir := filepath.Base(m.currentDir)

	// If we're in home directory, show ~
	if homeDir, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(m.currentDir, homeDir) {
			rel := strings.TrimPrefix(m.currentDir, homeDir)
			if rel == "" {
				dir = "~"
			} else {
				dir = "~" + rel
			}
		}
	}

	// Truncate if too long (shorter for statusline mode)
	// Keep the beginning (~/) and the project name
	if len(dir) > 50 {
		// For paths with ~, keep ~/ and last component
		if strings.HasPrefix(dir, "~/") {
			parts := strings.Split(dir, "/")
			if len(parts) >= 2 {
				lastPart := parts[len(parts)-1]
				// Allow up to 47 chars for ~/project-name
				if len(lastPart) > 47 {
					// Truncate the project name
					lastPart = "..." + lastPart[len(lastPart)-44:]
				}
				dir = "~/" + lastPart
			}
		} else {
			// For other paths, keep beginning
			dir = dir[:47] + "..."
		}
	}

	return dir
}

// FormatLanguageDisplay formats the language name with icon
func (m *Monitor) FormatLanguageDisplay() string {
	if m.language == "" {
		return ""
	}

	icon := getLanguageIcon(m.language)
	return fmt.Sprintf("%s %s", icon, m.language)
}

// getCPUUsage retrieves CPU usage on Linux/macOS
func getCPUUsage() (CPUInfo, error) {
	if runtime.GOOS == "linux" {
		return getLinuxCPUUsage()
	} else if runtime.GOOS == "darwin" {
		return getDarwinCPUUsage()
	}

	// Fallback: use 0
	return CPUInfo{CoreCount: runtime.NumCPU()}, nil
}

// getLinuxCPUUsage reads CPU usage from /proc/stat
func getLinuxCPUUsage() (CPUInfo, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return CPUInfo{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return CPUInfo{}, scanner.Err()
	}

	line := scanner.Text()
	fields := strings.Fields(line)

	if len(fields) < 8 || fields[0] != "cpu" {
		return CPUInfo{}, fmt.Errorf("invalid /proc/stat format")
	}

	// Parse CPU time fields
	// user, nice, system, idle, iowait, irq, softirq, steal
	user, _ := strconv.ParseFloat(fields[1], 64)
	nice, _ := strconv.ParseFloat(fields[2], 64)
	system, _ := strconv.ParseFloat(fields[3], 64)
	idle, _ := strconv.ParseFloat(fields[4], 64)

	total := user + nice + system + idle
	usage := total - idle

	var percent float64
	if total > 0 {
		percent = (usage / total) * 100
	}

	return CPUInfo{
		UsagePercent: percent,
		CoreCount:    runtime.NumCPU(),
	}, nil
}

// getDarwinCPUUsage reads CPU usage on macOS via sysctl
func getDarwinCPUUsage() (CPUInfo, error) {
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.thread_count")
	output, err := cmd.Output()
	if err != nil {
		return CPUInfo{}, err
	}

	cores, _ := strconv.Atoi(strings.TrimSpace(string(output)))

	// For simplicity on macOS, return 0 usage (would require more complex sysctl calls)
	return CPUInfo{
		UsagePercent: 0,
		CoreCount:    cores,
	}, nil
}

// getMemoryUsage retrieves memory usage
func getMemoryUsage() (MemoryInfo, error) {
	if runtime.GOOS == "linux" {
		return getLinuxMemoryUsage()
	} else if runtime.GOOS == "darwin" {
		return getDarwinMemoryUsage()
	}

	return MemoryInfo{}, nil
}

// getLinuxMemoryUsage reads memory info from /proc/meminfo
func getLinuxMemoryUsage() (MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryInfo{}, err
	}
	defer file.Close()

	var total, available uint64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(strings.TrimSuffix(fields[1], " kB"), 10, 64)
		if err != nil {
			continue
		}

		// Convert to bytes
		value = value * 1024

		switch fields[0] {
		case "MemTotal:":
			total = value
		case "MemAvailable:":
			available = value
		}
	}

	if total == 0 {
		return MemoryInfo{}, fmt.Errorf("could not determine total memory")
	}

	used := total - available
	percent := (float64(used) / float64(total)) * 100

	return MemoryInfo{
		Total:     total,
		Used:      used,
		Available: available,
		Percent:   percent,
	}, nil
}

// getDarwinMemoryUsage reads memory info on macOS
func getDarwinMemoryUsage() (MemoryInfo, error) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return MemoryInfo{}, err
	}

	total, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)

	// For simplicity, estimate available as 50% (would require vm_stat for accurate data)
	available := total / 2
	used := total - available
	percent := 50.0

	return MemoryInfo{
		Total:     total,
		Used:      used,
		Available: available,
		Percent:   percent,
	}, nil
}

// getDiskUsage retrieves disk usage for current partition
func getDiskUsage() (DiskInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return DiskInfo{}, err
	}

	var total, available uint64

	// Use df command for cross-platform compatibility
	cmd := exec.Command("df", "-k", cwd)
	output, err := cmd.Output()
	if err != nil {
		return DiskInfo{Path: cwd}, nil
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return DiskInfo{Path: cwd}, nil
	}

	// Parse df output
	// Skip header, get data line
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// fields[1] = total in KB, fields[3] = available in KB
		totalKB, err1 := strconv.ParseUint(fields[1], 10, 64)
		availKB, err2 := strconv.ParseUint(fields[3], 10, 64)

		if err1 == nil && err2 == nil {
			total = totalKB * 1024
			available = availKB * 1024
			break
		}
	}

	used := total - available
	var percent float64
	if total > 0 {
		percent = (float64(used) / float64(total)) * 100
	}

	return DiskInfo{
		Total:     total,
		Used:      used,
		Available: available,
		Percent:   percent,
		Path:      cwd,
	}, nil
}

// getFDCount retrieves the file descriptor count for the current process
func getFDCount() (FDInfo, error) {
	if runtime.GOOS == "linux" {
		return getLinuxFDCount()
	} else if runtime.GOOS == "darwin" {
		return getDarwinFDCount()
	}
	return FDInfo{}, nil
}

// getLinuxFDCount counts file descriptors by counting entries in /proc/self/fd
func getLinuxFDCount() (FDInfo, error) {
	fdPath := "/proc/self/fd"

	entries, err := os.ReadDir(fdPath)
	if err != nil {
		return FDInfo{}, err
	}

	return FDInfo{
		Count: len(entries),
	}, nil
}

// getDarwinFDCount counts file descriptors on macOS using lsof
func getDarwinFDCount() (FDInfo, error) {
	pid := os.Getpid()
	cmd := exec.Command("lsof", "-p", fmt.Sprintf("%d", pid))
	output, err := cmd.Output()
	if err != nil {
		return FDInfo{}, err
	}

	// Count lines (minus header)
	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "COMMAND") {
			count++
		}
	}

	return FDInfo{
		Count: count,
	}, nil
}

// detectLanguage detects the primary programming language from files in directory
func detectLanguage(dir string) string {
	// Language detection map based on file extensions
	extToLang := map[string]string{
		".go":  "Go",
		".py":  "Python",
		".rs":  "Rust",
		".js":  "JavaScript",
		".ts":  "TypeScript",
		".tsx": "TypeScript",
		".jsx": "JavaScript",
		".java": "Java",
		".kt":  "Kotlin",
		".rb":  "Ruby",
		".php": "PHP",
		".cs":  "C#",
		".cpp": "C++",
		".cc":  "C++",
		".cxx": "C++",
		".c":   "C",
		".h":   "C/C++",
		".hpp": "C++",
		".swift": "Swift",
		".sh":  "Shell",
		".scala": "Scala",
		".clj": "Clojure",
		".ex":  "Elixir",
		".exs": "Elixir",
		".erl": "Erlang",
		".hs":  "Haskell",
		".lua": "Lua",
		".r":   "R",
		".m":   "Objective-C",
		".ui":  "UI",
	}

	// Count files by extension
	langCounts := make(map[string]int)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden files and common directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if lang, ok := extToLang[ext]; ok {
			langCounts[lang]++
		}

		return nil
	})

	// Find most common language
	maxCount := 0
	detected := ""

	for lang, count := range langCounts {
		if count > maxCount {
			maxCount = count
			detected = lang
		}
	}

	return detected
}

// getLanguageIcon returns an icon for a programming language
func getLanguageIcon(lang string) string {
	icons := map[string]string{
		"Go":         "🐹",
		"Python":     "🐍",
		"Rust":       "🦀",
		"JavaScript": "🟨",
		"TypeScript": "💎",
		"Java":       "☕",
		"Ruby":       "💎",
		"C":          "🔧",
		"C++":        "⚙️",
		"C#":         "🔷",
		"Swift":      "🍎",
		"Shell":      "📜",
		"PHP":        "🐘",
		"Kotlin":     "🎯",
		"Elixir":     "💧",
		"Haskell":    "❓",
		"R":          "📊",
	}

	if icon, ok := icons[lang]; ok {
		return icon
	}

	return "📄"
}

// SetUpdateInterval sets how often to refresh metrics
func (m *Monitor) SetUpdateInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateInterval = interval
}

// ForceUpdate forces an immediate refresh of all metrics
func (m *Monitor) ForceUpdate() error {
	m.mu.Lock()
	m.lastUpdate = time.Time{}
	m.mu.Unlock()
	return m.Update()
}
