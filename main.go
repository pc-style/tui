package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type tickMsg time.Time

type processData struct {
	processes    []Process
	systemStats  SystemStats
	lastUpdate   time.Time
}

type Process struct {
	PID         int
	Name        string
	User        string
	CPU         float64
	Memory      float64
	State       string
	PPID        int
	StartTime   string
	Command     string
}

type SystemStats struct {
	CPUCores    int
	CPUUsage    []float64
	MemTotal    uint64
	MemUsed     uint64
	SwapTotal   uint64
	SwapUsed    uint64
	LoadAvg     [3]float64
	Uptime      time.Duration
}

type NetworkConnection struct {
	Protocol string
	LocalAddr string
	LocalPort string
	RemoteAddr string
	RemotePort string
	State string
	PID int
	ProcessName string
}

type NetworkInterface struct {
	Name string
	BytesIn uint64
	BytesOut uint64
	PacketsIn uint64
	PacketsOut uint64
	Speed uint64
}

type SystemAlert struct {
	Type string
	Level string // INFO, WARN, ERROR, CRITICAL
	Message string
	Timestamp time.Time
	Source string
}

type model struct {
	width       int
	height      int
	frame       int
	quitting    bool
	err         error
	
	// Current view mode
	viewMode    string // "processes", "network", "alerts", "overview"
	
	// Process management
	processes     []Process
	systemStats   SystemStats
	selectedRow   int
	sortColumn    string
	sortReverse   bool
	searchMode    bool
	searchQuery   string
	filteredProcs []Process
	scrollOffset  int
	
	// Network monitoring
	connections []NetworkConnection
	interfaces []NetworkInterface
	networkTraffic []float64 // Recent traffic history
	bandwidthIn []float64
	bandwidthOut []float64
	
	// System alerts
	alerts []SystemAlert
	alertScrollOffset int
	
	// UI state
	showKillDialog bool
	killPID        int
	selectedPanel  int // 0=left, 1=center, 2=right
	
	// Visual effects
	particles     []particle
	cpuHistory    [][]float64
	memHistory    []float64
	networkWaves  []networkWave
}

type particle struct {
	x, y   float64
	vx, vy float64
	char   rune
	color  string
}

type networkWave struct {
	x, y   float64
	amplitude float64
	frequency float64
	phase float64
	color string
}

var (
	neonCyan    = lipgloss.Color("#00FFFF")
	neonMagenta = lipgloss.Color("#FF00FF") 
	neonGreen   = lipgloss.Color("#00FF00")
	neonYellow  = lipgloss.Color("#FFFF00")
	neonRed     = lipgloss.Color("#FF0066")
	darkBg      = lipgloss.Color("#0D0D0D")
)

var (
	quitKeys = key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("", "press q to quit"),
	)
	
	killKeys = key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("", "kill process"),
	)
	
	searchKeys = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("", "search"),
	)
	
	tabKeys = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("", "switch view"),
	)
)

func initialModel() model {
	numCores := runtime.NumCPU()
	m := model{
		viewMode:     "overview", // Start with overview
		particles:    make([]particle, 20),
		networkWaves: make([]networkWave, 10),
		cpuHistory:   make([][]float64, numCores),
		memHistory:   make([]float64, 50),
		bandwidthIn:  make([]float64, 50),
		bandwidthOut: make([]float64, 50),
		networkTraffic: make([]float64, 50),
		sortColumn:   "CPU",
		sortReverse:  true,
		selectedRow:  0,
		scrollOffset: 0,
		selectedPanel: 0,
	}
	
	// Initialize CPU history for each core
	for i := range m.cpuHistory {
		m.cpuHistory[i] = make([]float64, 50)
	}
	
	// Initialize floating particles
	for i := range m.particles {
		m.particles[i] = particle{
			x:     rand.Float64() * 100,
			y:     rand.Float64() * 30,
			vx:    (rand.Float64() - 0.5) * 0.3,
			vy:    (rand.Float64() - 0.5) * 0.2,
			char:  []rune("•◦◊◦•○●◎◉▲▼►◄")[rand.Intn(12)],
			color: []string{string(neonCyan), string(neonMagenta), string(neonGreen), string(neonYellow)}[rand.Intn(4)],
		}
	}
	
	// Initialize network waves
	for i := range m.networkWaves {
		m.networkWaves[i] = networkWave{
			x:         rand.Float64() * 100,
			y:         rand.Float64() * 30,
			amplitude: rand.Float64() * 3 + 1,
			frequency: rand.Float64() * 0.5 + 0.1,
			phase:     rand.Float64() * 6.28, // 2π
			color:     []string{string(neonCyan), string(neonMagenta), string(neonGreen)}[rand.Intn(3)],
		}
	}
	
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tick(),
		fetchProcesses(),
		fetchNetworkData(),
		fetchSystemAlerts(),
	)
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchProcesses() tea.Cmd {
	return func() tea.Msg {
		processes, stats := getProcessesAndStats()
		return processData{
			processes:   processes,
			systemStats: stats,
			lastUpdate:  time.Now(),
		}
	}
}

type networkData struct {
	connections []NetworkConnection
	interfaces []NetworkInterface
	lastUpdate time.Time
}

type alertData struct {
	alerts []SystemAlert
	lastUpdate time.Time
}

func fetchNetworkData() tea.Cmd {
	return func() tea.Msg {
		connections := getNetworkConnections()
		interfaces := getNetworkInterfaces()
		return networkData{
			connections: connections,
			interfaces: interfaces,
			lastUpdate: time.Now(),
		}
	}
}

func fetchSystemAlerts() tea.Cmd {
	return func() tea.Msg {
		alerts := getSystemAlerts()
		return alertData{
			alerts: alerts,
			lastUpdate: time.Now(),
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		if m.showKillDialog {
			switch msg.String() {
			case "y", "Y":
				killProcess(m.killPID)
				m.showKillDialog = false
				return m, fetchProcesses()
			case "n", "N", "esc":
				m.showKillDialog = false
				return m, nil
			}
			return m, nil
		}
		
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.filteredProcs = m.processes
				return m, nil
			case "enter":
				m.searchMode = false
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filteredProcs = filterProcesses(m.processes, m.searchQuery)
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.filteredProcs = filterProcesses(m.processes, m.searchQuery)
				}
				return m, nil
			}
		}
		
		switch {
		case key.Matches(msg, quitKeys):
			m.quitting = true
			return m, tea.Quit
			
		case key.Matches(msg, tabKeys):
			// Cycle through view modes
			modes := []string{"overview", "processes", "network", "alerts"}
			currentIndex := 0
			for i, mode := range modes {
				if mode == m.viewMode {
					currentIndex = i
					break
				}
			}
			m.viewMode = modes[(currentIndex+1)%len(modes)]
			m.selectedRow = 0
			m.scrollOffset = 0
			return m, nil
			
		case msg.String() == "1":
			m.viewMode = "overview"
			return m, nil
		case msg.String() == "2":
			m.viewMode = "processes"
			return m, nil
		case msg.String() == "3":
			m.viewMode = "network"
			return m, nil
		case msg.String() == "4":
			m.viewMode = "alerts"
			return m, nil
			
		case key.Matches(msg, killKeys):
			if m.viewMode == "processes" && len(m.filteredProcs) > 0 && m.selectedRow < len(m.filteredProcs) {
				m.killPID = m.filteredProcs[m.selectedRow].PID
				m.showKillDialog = true
			}
			return m, nil
			
		case key.Matches(msg, searchKeys):
			if m.viewMode == "processes" {
				m.searchMode = true
				m.searchQuery = ""
			}
			return m, nil
			
		case msg.String() == "up", msg.String() == "k":
			if m.selectedRow > 0 {
				m.selectedRow--
				if m.selectedRow < m.scrollOffset {
					m.scrollOffset = m.selectedRow
				}
			}
			return m, nil
			
		case msg.String() == "down", msg.String() == "j":
			maxItems := 0
			switch m.viewMode {
			case "processes":
				maxItems = len(m.filteredProcs)
			case "network":
				maxItems = len(m.connections)
			case "alerts":
				maxItems = len(m.alerts)
			}
			
			if m.selectedRow < maxItems-1 {
				m.selectedRow++
				visibleRows := m.height - 12
				if m.selectedRow >= m.scrollOffset+visibleRows {
					m.scrollOffset = m.selectedRow - visibleRows + 1
				}
			}
			return m, nil
			
		case msg.String() == "c":
			if m.viewMode == "processes" {
				m.sortColumn = "CPU"
				m.sortProcesses()
			}
			return m, nil
			
		case msg.String() == "m":
			if m.viewMode == "processes" {
				m.sortColumn = "Memory"
				m.sortProcesses()
			}
			return m, nil
			
		case msg.String() == "p":
			if m.viewMode == "processes" {
				m.sortColumn = "PID"
				m.sortProcesses()
			}
			return m, nil
		}
		
		return m, nil
		
	case tickMsg:
		m.frame++
		
		// Update particles
		for i := range m.particles {
			p := &m.particles[i]
			p.x += p.vx
			p.y += p.vy
			
			if p.x < 0 || p.x > 100 {
				p.vx *= -1
			}
			if p.y < 0 || p.y > 30 {
				p.vy *= -1
			}
		}
		
		// Update network waves
		for i := range m.networkWaves {
			w := &m.networkWaves[i]
			w.phase += 0.1
			w.x += 0.5
			if w.x > 100 {
				w.x = 0
			}
		}
		
		// Fetch new data periodically
		var cmds []tea.Cmd
		if m.frame%25 == 0 { // ~2 seconds
			cmds = append(cmds, fetchProcesses())
		}
		if m.frame%30 == 0 { // ~2.4 seconds  
			cmds = append(cmds, fetchNetworkData())
		}
		if m.frame%50 == 0 { // ~4 seconds
			cmds = append(cmds, fetchSystemAlerts())
		}
		
		cmds = append(cmds, tick())
		return m, tea.Batch(cmds...)
		
	case processData:
		m.processes = msg.processes
		m.systemStats = msg.systemStats
		
		// Update CPU history
		for i, usage := range msg.systemStats.CPUUsage {
			if i < len(m.cpuHistory) {
				m.cpuHistory[i] = append(m.cpuHistory[i][1:], usage)
			}
		}
		
		// Update memory history
		memUsagePercent := float64(msg.systemStats.MemUsed) / float64(msg.systemStats.MemTotal) * 100
		m.memHistory = append(m.memHistory[1:], memUsagePercent)
		
		// Sort and filter processes
		m.sortProcesses()
		m.filteredProcs = filterProcesses(m.processes, m.searchQuery)
		
		// Ensure selected row is valid
		if m.selectedRow >= len(m.filteredProcs) {
			m.selectedRow = len(m.filteredProcs) - 1
		}
		if m.selectedRow < 0 {
			m.selectedRow = 0
		}
		
		return m, nil
		
	case networkData:
		m.connections = msg.connections
		m.interfaces = msg.interfaces
		
		// Update bandwidth history (simulate for now)
		totalIn := 0.0
		totalOut := 0.0
		for _, iface := range msg.interfaces {
			totalIn += float64(iface.BytesIn)
			totalOut += float64(iface.BytesOut)
		}
		
		m.bandwidthIn = append(m.bandwidthIn[1:], totalIn/1024/1024) // MB
		m.bandwidthOut = append(m.bandwidthOut[1:], totalOut/1024/1024) // MB
		m.networkTraffic = append(m.networkTraffic[1:], (totalIn+totalOut)/1024/1024)
		
		return m, nil
		
	case alertData:
		// Add new alerts to the beginning
		for _, alert := range msg.alerts {
			m.alerts = append([]SystemAlert{alert}, m.alerts...)
		}
		
		// Keep only last 100 alerts
		if len(m.alerts) > 100 {
			m.alerts = m.alerts[:100]
		}
		
		return m, nil
		
	case errMsg:
		m.err = msg
		return m, nil
		
	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	
	if m.width == 0 || m.height == 0 {
		return "Initializing CYBERPUNK COMMAND CENTER..."
	}
	
	var s strings.Builder
	
	// Dynamic header based on current view
	var title string
	switch m.viewMode {
	case "overview":
		title = "⚡ CYBERPUNK COMMAND CENTER v3.0 ⚡"
	case "processes":
		title = "⚡ PROCESS MONITOR ⚡"
	case "network":
		title = "🌐 NETWORK ANALYZER 🌐"
	case "alerts":
		title = "🚨 SYSTEM ALERTS 🚨"
	default:
		title = "⚡ COMMAND CENTER ⚡"
	}
	
	// Animated glitch effects
	glitchTitle := title
	if m.frame%40 < 10 {
		glitchTitle = strings.ReplaceAll(title, "C", "Ɔ")
		glitchTitle = strings.ReplaceAll(glitchTitle, "E", "Ξ")
		glitchTitle = strings.ReplaceAll(glitchTitle, "R", "Я")
	}
	
	headerStyle := lipgloss.NewStyle().
		Foreground(neonCyan).
		Background(darkBg).
		Bold(true).
		Padding(0, 1).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(neonMagenta).
		Width(m.width - 2).
		Align(lipgloss.Center)
	
	s.WriteString(headerStyle.Render(glitchTitle))
	s.WriteString("\n")
	
	// View mode tabs
	s.WriteString(m.renderViewTabs())
	s.WriteString("\n")
	
	// Main content based on view mode
	switch m.viewMode {
	case "overview":
		s.WriteString(m.renderOverviewDashboard())
	case "processes":
		s.WriteString(m.renderProcessView())
	case "network":
		s.WriteString(m.renderNetworkView())
	case "alerts":
		s.WriteString(m.renderAlertsView())
	}
	
	// Controls footer
	s.WriteString(m.renderMainControls())
	
	// Kill dialog
	if m.showKillDialog {
		s.WriteString(m.renderKillDialog())
	}
	
	if m.quitting {
		s.WriteString("\n" + lipgloss.NewStyle().
			Foreground(neonRed).
			Bold(true).
			Align(lipgloss.Center).
			Width(m.width).
			Render(">>> SHUTTING DOWN NEURAL MATRIX... <<<"))
	}
	
	return s.String()
}

// Network data fetching
func getNetworkConnections() []NetworkConnection {
	connections := []NetworkConnection{}
	
	// Get network connections using netstat
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return connections
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue // Skip headers
		}
		
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		protocol := fields[0]
		localAddr := fields[3]
		var remoteAddr, state string
		
		if len(fields) >= 5 {
			remoteAddr = fields[4]
		}
		if len(fields) >= 6 {
			state = fields[5]
		}
		
		// Parse addresses
		localParts := strings.Split(localAddr, ":")
		remoteParts := strings.Split(remoteAddr, ":")
		
		conn := NetworkConnection{
			Protocol: protocol,
			State: state,
		}
		
		if len(localParts) >= 2 {
			conn.LocalAddr = strings.Join(localParts[:len(localParts)-1], ":")
			conn.LocalPort = localParts[len(localParts)-1]
		}
		
		if len(remoteParts) >= 2 {
			conn.RemoteAddr = strings.Join(remoteParts[:len(remoteParts)-1], ":")
			conn.RemotePort = remoteParts[len(remoteParts)-1]
		}
		
		connections = append(connections, conn)
	}
	
	return connections
}

func getNetworkInterfaces() []NetworkInterface {
	interfaces := []NetworkInterface{}
	
	// Simulate network interface data for now
	interfaces = append(interfaces, NetworkInterface{
		Name: "en0",
		BytesIn: uint64(rand.Intn(1000000000)),
		BytesOut: uint64(rand.Intn(1000000000)),
		PacketsIn: uint64(rand.Intn(1000000)),
		PacketsOut: uint64(rand.Intn(1000000)),
		Speed: 1000000000, // 1Gbps
	})
	
	return interfaces
}

func getSystemAlerts() []SystemAlert {
	alerts := []SystemAlert{}
	
	// Generate some sample alerts
	alertTypes := []string{"CPU", "Memory", "Disk", "Network", "Process"}
	levels := []string{"INFO", "WARN", "ERROR"}
	
	if rand.Float64() < 0.3 { // 30% chance of new alert
		alertType := alertTypes[rand.Intn(len(alertTypes))]
		level := levels[rand.Intn(len(levels))]
		
		var message string
		switch alertType {
		case "CPU":
			message = fmt.Sprintf("CPU usage is %.1f%%", rand.Float64()*100)
		case "Memory":
			message = fmt.Sprintf("Memory usage is %.1f%%", rand.Float64()*100)
		case "Network":
			message = "High network traffic detected"
		default:
			message = fmt.Sprintf("%s alert triggered", alertType)
		}
		
		alerts = append(alerts, SystemAlert{
			Type: alertType,
			Level: level,
			Message: message,
			Timestamp: time.Now(),
			Source: "system",
		})
	}
	
	return alerts
}

// Process management functions
func getProcessesAndStats() ([]Process, SystemStats) {
	processes := []Process{}
	stats := SystemStats{
		CPUCores: runtime.NumCPU(),
		CPUUsage: make([]float64, runtime.NumCPU()),
	}
	
	// Get processes using ps command
	cmd := exec.Command("ps", "ax", "-o", "pid,user,pcpu,pmem,stat,ppid,etime,comm")
	output, err := cmd.Output()
	if err != nil {
		return processes, stats
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}
		
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		
		pid, _ := strconv.Atoi(fields[0])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		ppid, _ := strconv.Atoi(fields[5])
		
		process := Process{
			PID:       pid,
			User:      fields[1],
			CPU:       cpu,
			Memory:    mem,
			State:     fields[4],
			PPID:      ppid,
			StartTime: fields[6],
			Name:      fields[7],
			Command:   strings.Join(fields[7:], " "),
		}
		processes = append(processes, process)
	}
	
	// Get memory info
	cmd = exec.Command("sysctl", "-n", "hw.memsize")
	if output, err := cmd.Output(); err == nil {
		if total, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64); err == nil {
			stats.MemTotal = total
		}
	}
	
	// Get memory usage (simplified)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.MemUsed = m.Alloc
	
	// Simulate CPU usage per core
	for i := range stats.CPUUsage {
		stats.CPUUsage[i] = math.Max(0, math.Min(100, math.Sin(float64(time.Now().UnixNano()/1000000000+int64(i)))*30+50+rand.Float64()*20))
	}
	
	// Get load average
	cmd = exec.Command("sysctl", "-n", "vm.loadavg")
	if output, err := cmd.Output(); err == nil {
		parts := strings.Fields(strings.TrimSpace(string(output)))
		if len(parts) >= 3 {
			stats.LoadAvg[0], _ = strconv.ParseFloat(parts[1], 64)
			stats.LoadAvg[1], _ = strconv.ParseFloat(parts[2], 64)
			stats.LoadAvg[2], _ = strconv.ParseFloat(parts[3], 64)
		}
	}
	
	return processes, stats
}

func killProcess(pid int) {
	syscall.Kill(pid, syscall.SIGTERM)
}

func (m *model) sortProcesses() {
	sort.Slice(m.processes, func(i, j int) bool {
		switch m.sortColumn {
		case "CPU":
			if m.sortReverse {
				return m.processes[i].CPU > m.processes[j].CPU
			}
			return m.processes[i].CPU < m.processes[j].CPU
		case "Memory":
			if m.sortReverse {
				return m.processes[i].Memory > m.processes[j].Memory
			}
			return m.processes[i].Memory < m.processes[j].Memory
		case "PID":
			if m.sortReverse {
				return m.processes[i].PID > m.processes[j].PID
			}
			return m.processes[i].PID < m.processes[j].PID
		default:
			return m.processes[i].PID < m.processes[j].PID
		}
	})
}

func filterProcesses(processes []Process, query string) []Process {
	if query == "" {
		return processes
	}
	
	filtered := []Process{}
	query = strings.ToLower(query)
	
	for _, p := range processes {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.User), query) ||
			strings.Contains(strings.ToLower(p.Command), query) {
			filtered = append(filtered, p)
		}
	}
	
	return filtered
}

// Render functions
func (m model) renderSystemOverview() string {
	var s strings.Builder
	
	// CPU cores overview
	cpuSection := lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true).
		Render("⚡ CPU CORES")
	
	s.WriteString(cpuSection + "\n")
	
	coresPerRow := 4
	for i := 0; i < len(m.systemStats.CPUUsage); i += coresPerRow {
		for j := 0; j < coresPerRow && i+j < len(m.systemStats.CPUUsage); j++ {
			idx := i + j
			usage := 0.0
			if idx < len(m.systemStats.CPUUsage) {
				usage = m.systemStats.CPUUsage[idx]
			}
			
			bar := renderBar(usage, 100, 10, neonGreen, neonRed)
			coreLabel := fmt.Sprintf("CPU%d", idx)
			
			color := neonGreen
			if usage > 80 {
				color = neonRed
			} else if usage > 60 {
				color = neonYellow
			}
			
			s.WriteString(lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%-5s %s %.1f%%", coreLabel, bar, usage)))
			if j < coresPerRow-1 && i+j+1 < len(m.systemStats.CPUUsage) {
				s.WriteString("  ")
			}
		}
		s.WriteString("\n")
	}
	
	// Memory and Load info
	memUsage := float64(m.systemStats.MemUsed) / float64(m.systemStats.MemTotal) * 100
	memBar := renderBar(memUsage, 100, 30, neonCyan, neonRed)
	
	loadAvgColor := neonGreen
	if m.systemStats.LoadAvg[0] > float64(m.systemStats.CPUCores) {
		loadAvgColor = neonRed
	}
	
	memInfo := fmt.Sprintf("MEM: %s %.1f%% (%s / %s)",
		memBar, memUsage,
		formatBytes(m.systemStats.MemUsed),
		formatBytes(m.systemStats.MemTotal))
	
	loadInfo := fmt.Sprintf("Load: %.2f %.2f %.2f",
		m.systemStats.LoadAvg[0], m.systemStats.LoadAvg[1], m.systemStats.LoadAvg[2])
	
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(neonCyan).Render(memInfo) + "\n")
	s.WriteString(lipgloss.NewStyle().Foreground(loadAvgColor).Render(loadInfo) + "\n")
	
	return s.String()
}

func (m model) renderProcessList() string {
	var s strings.Builder
	
	// Process list header
	headerText := fmt.Sprintf("⚡ PROCESSES (%d) - Sort: %s", len(m.filteredProcs), m.sortColumn)
	if m.searchQuery != "" {
		headerText += fmt.Sprintf(" - Search: '%s'", m.searchQuery)
	}
	
	header := lipgloss.NewStyle().
		Foreground(neonMagenta).
		Bold(true).
		Render(headerText)
	
	s.WriteString(header + "\n")
	
	// Column headers
	pidHeader := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Width(8).Render("PID")
	userHeader := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Width(12).Render("USER")
	cpuHeader := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Width(8).Render("CPU%")
	memHeader := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Width(8).Render("MEM%")
	stateHeader := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Width(8).Render("STATE")
	nameHeader := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Render("COMMAND")
	
	headerRow := fmt.Sprintf("%s %s %s %s %s %s", 
		pidHeader, userHeader, cpuHeader, memHeader, stateHeader, nameHeader)
	
	s.WriteString(headerRow + "\n")
	s.WriteString(strings.Repeat("─", m.width-4) + "\n")
	
	// Process rows
	visibleRows := m.height - 12 // Account for headers and footers
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleRows
	
	if endIdx > len(m.filteredProcs) {
		endIdx = len(m.filteredProcs)
	}
	
	for i := startIdx; i < endIdx; i++ {
		if i >= len(m.filteredProcs) {
			break
		}
		
		proc := m.filteredProcs[i]
		isSelected := i == m.selectedRow
		
		// Process state colors
		stateColor := neonGreen
		switch proc.State {
		case "R", "R+":
			stateColor = neonGreen
		case "S", "S+":
			stateColor = neonCyan
		case "T":
			stateColor = neonYellow
		case "Z":
			stateColor = neonRed
		}
		
		// CPU color based on usage
		cpuColor := neonGreen
		if proc.CPU > 50 {
			cpuColor = neonRed
		} else if proc.CPU > 25 {
			cpuColor = neonYellow
		}
		
		// Memory color
		memColor := neonGreen
		if proc.Memory > 5 {
			memColor = neonRed
		} else if proc.Memory > 2 {
			memColor = neonYellow
		}
		
		// Truncate command if too long
		command := proc.Command
		maxCmdLen := m.width - 50
		if len(command) > maxCmdLen {
			command = command[:maxCmdLen-3] + "..."
		}
		
		pidText := lipgloss.NewStyle().Width(8).Render(fmt.Sprintf("%d", proc.PID))
		userText := lipgloss.NewStyle().Width(12).Render(proc.User)
		cpuText := lipgloss.NewStyle().Foreground(cpuColor).Width(8).Render(fmt.Sprintf("%.1f", proc.CPU))
		memText := lipgloss.NewStyle().Foreground(memColor).Width(8).Render(fmt.Sprintf("%.1f", proc.Memory))
		stateText := lipgloss.NewStyle().Foreground(stateColor).Width(8).Render(proc.State)
		cmdText := lipgloss.NewStyle().Render(command)
		
		row := fmt.Sprintf("%s %s %s %s %s %s", 
			pidText, userText, cpuText, memText, stateText, cmdText)
		
		if isSelected {
			row = lipgloss.NewStyle().
				Foreground(darkBg).
				Background(neonCyan).
				Bold(true).
				Width(m.width-4).
				Render(row)
		}
		
		s.WriteString(row + "\n")
	}
	
	// Add particles as background effect
	if len(m.particles) > 0 {
		particleLines := make([]string, 3)
		for _, p := range m.particles[:min(len(m.particles), 10)] {
			if int(p.y)%3 < len(particleLines) {
				particleLines[int(p.y)%3] += string(p.char)
			}
		}
		for _, line := range particleLines {
			if line != "" {
				s.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.particles[0].color)).
					Faint(true).
					Render(line) + "\n")
			}
		}
	}
	
	return s.String()
}

func (m model) renderControls() string {
	controls := []string{}
	
	if m.searchMode {
		controls = append(controls, 
			lipgloss.NewStyle().Foreground(neonYellow).Render("SEARCH: "+m.searchQuery+"_"),
			lipgloss.NewStyle().Foreground(neonCyan).Render("ESC: exit search"),
		)
	} else {
		controls = append(controls,
			lipgloss.NewStyle().Foreground(neonGreen).Render("↑/↓: navigate"),
			lipgloss.NewStyle().Foreground(neonMagenta).Render("k: kill"),
			lipgloss.NewStyle().Foreground(neonYellow).Render("/: search"),
			lipgloss.NewStyle().Foreground(neonCyan).Render("c/m/p: sort"),
			lipgloss.NewStyle().Foreground(neonRed).Render("q: quit"),
		)
	}
	
	return "\n" + lipgloss.NewStyle().
		Background(darkBg).
		Padding(0, 1).
		Width(m.width).
		Render(strings.Join(controls, " | "))
}

func (m model) renderKillDialog() string {
	if !m.showKillDialog {
		return ""
	}
	
	dialog := fmt.Sprintf("Kill process %d? (y/N)", m.killPID)
	
	style := lipgloss.NewStyle().
		Foreground(neonRed).
		Background(darkBg).
		Bold(true).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(neonRed).
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(30)
	
	// Center the dialog
	return "\n" + lipgloss.Place(m.width, 5, lipgloss.Center, lipgloss.Center, style.Render(dialog))
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// New render functions for the command center
func (m model) renderViewTabs() string {
	tabs := []string{"1:Overview", "2:Processes", "3:Network", "4:Alerts"}
	activeIndex := 0
	
	switch m.viewMode {
	case "overview": activeIndex = 0
	case "processes": activeIndex = 1
	case "network": activeIndex = 2
	case "alerts": activeIndex = 3
	}
	
	var tabStrs []string
	for i, tab := range tabs {
		if i == activeIndex {
			tabStrs = append(tabStrs, lipgloss.NewStyle().
				Foreground(darkBg).
				Background(neonCyan).
				Bold(true).
				Padding(0, 1).
				Render(tab))
		} else {
			tabStrs = append(tabStrs, lipgloss.NewStyle().
				Foreground(neonCyan).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(neonCyan).
				Padding(0, 1).
				Render(tab))
		}
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, tabStrs...)
}

func (m model) renderOverviewDashboard() string {
	// Create three-panel overview
	leftPanel := m.renderSystemOverview()
	centerPanel := m.renderNetworkOverview()
	rightPanel := m.renderAlertsOverview()
	
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Height(m.height - 10)
	
	leftStyled := panelStyle.Copy().
		BorderForeground(neonCyan).
		Width(m.width/3 - 2).
		Render(leftPanel)
	
	centerStyled := panelStyle.Copy().
		BorderForeground(neonMagenta).
		Width(m.width/3 - 2).
		Render(centerPanel)
	
	rightStyled := panelStyle.Copy().
		BorderForeground(neonYellow).
		Width(m.width/3 - 2).
		Render(rightPanel)
	
	return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, centerStyled, rightStyled)
}

func (m model) renderProcessView() string {
	return m.renderSystemOverview() + "\n" + m.renderProcessList()
}

func (m model) renderNetworkView() string {
	var s strings.Builder
	
	s.WriteString(lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true).
		Render("🌐 NETWORK CONNECTIONS\n\n"))
	
	// Network interfaces summary
	for _, iface := range m.interfaces {
		s.WriteString(fmt.Sprintf("Interface: %s\n", iface.Name))
		s.WriteString(fmt.Sprintf("  In:  %s\n", formatBytes(iface.BytesIn)))
		s.WriteString(fmt.Sprintf("  Out: %s\n\n", formatBytes(iface.BytesOut)))
	}
	
	// Active connections
	s.WriteString(lipgloss.NewStyle().
		Foreground(neonMagenta).
		Bold(true).
		Render("ACTIVE CONNECTIONS:\n"))
	
	maxRows := m.height - 15
	for i, conn := range m.connections {
		if i >= maxRows {
			break
		}
		
		stateColor := neonGreen
		if conn.State == "ESTABLISHED" {
			stateColor = neonGreen
		} else if conn.State == "LISTEN" {
			stateColor = neonCyan
		} else {
			stateColor = neonYellow
		}
		
		line := fmt.Sprintf("%-6s %s:%s -> %s:%s [%s]",
			conn.Protocol, conn.LocalAddr, conn.LocalPort,
			conn.RemoteAddr, conn.RemotePort, conn.State)
		
		s.WriteString(lipgloss.NewStyle().Foreground(stateColor).Render(line) + "\n")
	}
	
	return s.String()
}

func (m model) renderAlertsView() string {
	var s strings.Builder
	
	s.WriteString(lipgloss.NewStyle().
		Foreground(neonRed).
		Bold(true).
		Render("🚨 SYSTEM ALERTS\n\n"))
	
	if len(m.alerts) == 0 {
		s.WriteString(lipgloss.NewStyle().
			Foreground(neonGreen).
			Render("✓ All systems operational"))
		return s.String()
	}
	
	maxRows := m.height - 10
	for i, alert := range m.alerts {
		if i >= maxRows {
			break
		}
		
		var levelColor lipgloss.Color
		switch alert.Level {
		case "ERROR":
			levelColor = neonRed
		case "WARN":
			levelColor = neonYellow
		case "INFO":
			levelColor = neonGreen
		default:
			levelColor = neonCyan
		}
		
		timestamp := alert.Timestamp.Format("15:04:05")
		line := fmt.Sprintf("[%s] %s: %s (%s)",
			timestamp, alert.Level, alert.Message, alert.Type)
		
		s.WriteString(lipgloss.NewStyle().Foreground(levelColor).Render(line) + "\n")
	}
	
	return s.String()
}

func (m model) renderNetworkOverview() string {
	var s strings.Builder
	
	s.WriteString(lipgloss.NewStyle().
		Foreground(neonMagenta).
		Bold(true).
		Render("🌐 NETWORK STATUS\n\n"))
	
	// Traffic overview
	if len(m.networkTraffic) > 0 {
		traffic := m.networkTraffic[len(m.networkTraffic)-1]
		s.WriteString(fmt.Sprintf("Total Traffic: %.2f MB/s\n", traffic))
	}
	
	s.WriteString(fmt.Sprintf("Active Connections: %d\n", len(m.connections)))
	s.WriteString(fmt.Sprintf("Network Interfaces: %d\n\n", len(m.interfaces)))
	
	// Network waves animation
	s.WriteString("Network Activity:\n")
	for i := 0; i < 8; i++ {
		line := ""
		for j := 0; j < 30; j++ {
			char := "·"
			for _, wave := range m.networkWaves {
				if int(wave.x)%30 == j {
					waveY := int(wave.y + wave.amplitude*math.Sin(wave.phase+float64(j)*wave.frequency))
					if waveY%8 == i {
						char = "~"
						break
					}
				}
			}
			line += char
		}
		s.WriteString(lipgloss.NewStyle().
			Foreground(neonCyan).
			Render(line) + "\n")
	}
	
	return s.String()
}

func (m model) renderAlertsOverview() string {
	var s strings.Builder
	
	s.WriteString(lipgloss.NewStyle().
		Foreground(neonYellow).
		Bold(true).
		Render("⚠ RECENT ALERTS\n\n"))
	
	if len(m.alerts) == 0 {
		s.WriteString(lipgloss.NewStyle().
			Foreground(neonGreen).
			Render("✓ No active alerts"))
		return s.String()
	}
	
	// Show last 10 alerts
	maxAlerts := min(10, len(m.alerts))
	for i := 0; i < maxAlerts; i++ {
		alert := m.alerts[i]
		
		var levelColor lipgloss.Color
		var icon string
		switch alert.Level {
		case "ERROR":
			levelColor = neonRed
			icon = "❌"
		case "WARN":
			levelColor = neonYellow
			icon = "⚠"
		case "INFO":
			levelColor = neonGreen
			icon = "ℹ"
		default:
			levelColor = neonCyan
			icon = "·"
		}
		
		time := alert.Timestamp.Format("15:04")
		line := fmt.Sprintf("%s %s %s", icon, time, alert.Message)
		if len(line) > 35 {
			line = line[:32] + "..."
		}
		
		s.WriteString(lipgloss.NewStyle().Foreground(levelColor).Render(line) + "\n")
	}
	
	return s.String()
}

func (m model) renderMainControls() string {
	controls := []string{}
	
	switch m.viewMode {
	case "processes":
		if m.searchMode {
			controls = append(controls, 
				lipgloss.NewStyle().Foreground(neonYellow).Render("SEARCH: "+m.searchQuery+"_"),
				lipgloss.NewStyle().Foreground(neonCyan).Render("ESC: exit search"),
			)
		} else {
			controls = append(controls,
				lipgloss.NewStyle().Foreground(neonGreen).Render("↑/↓: navigate"),
				lipgloss.NewStyle().Foreground(neonMagenta).Render("k: kill"),
				lipgloss.NewStyle().Foreground(neonYellow).Render("/: search"),
				lipgloss.NewStyle().Foreground(neonCyan).Render("c/m/p: sort"),
			)
		}
	default:
		controls = append(controls,
			lipgloss.NewStyle().Foreground(neonGreen).Render("TAB: switch view"),
			lipgloss.NewStyle().Foreground(neonCyan).Render("1-4: quick switch"),
		)
	}
	
	controls = append(controls, lipgloss.NewStyle().Foreground(neonRed).Render("q: quit"))
	
	return "\n" + lipgloss.NewStyle().
		Background(darkBg).
		Padding(0, 1).
		Width(m.width).
		Render(strings.Join(controls, " | "))
}

func renderBar(value, max float64, width int, colorLow, colorHigh lipgloss.Color) string {
	filled := int((value / max) * float64(width))
	if filled > width {
		filled = width
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	color := colorLow
	if value/max > 0.7 {
		color = colorHigh
	}
	
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}



func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
