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

type processData struct {
	processes   []Process
	systemStats SystemStats
	lastUpdate  time.Time
}

type model struct {
	width    int
	height   int
	frame    int
	quitting bool
	err      error

	viewMode string

	processes     []Process
	systemStats   SystemStats
	selectedRow   int
	sortColumn    string
	sortReverse   bool
	searchMode    bool
	searchQuery   string
	filteredProcs []Process
	scrollOffset  int

	connections    []NetworkConnection
	interfaces     []NetworkInterface
	networkTraffic []float64
	bandwidthIn    []float64
	bandwidthOut   []float64

	alerts            []SystemAlert
	alertScrollOffset int

	showKillDialog bool
	killPID        int
	selectedPanel  int

	// Enhanced visual effects
	orbs          []orb
	lightRays     []lightRay
	matrixRain    []matrixDrop
	energyField   []energyNode
	cpuHistory    [][]float64
	memHistory    []float64
	glowIntensity float64
	pulsePhase    float64
}

// Elegant color palette
var (
	// Primary colors - more subtle and elegant
	primary   = lipgloss.Color("#6366f1") // Indigo
	secondary = lipgloss.Color("#8b5cf6") // Violet
	accent    = lipgloss.Color("#06b6d4") // Cyan
	success   = lipgloss.Color("#10b981") // Emerald
	warning   = lipgloss.Color("#f59e0b") // Amber
	danger    = lipgloss.Color("#ef4444") // Red

	// Background and text
	bg        = lipgloss.Color("#0f0f23") // Deep space blue
	bgLight   = lipgloss.Color("#1e1e3f") // Lighter background
	text      = lipgloss.Color("#e2e8f0") // Light gray
	textMuted = lipgloss.Color("#94a3b8") // Muted gray
	textFaint = lipgloss.Color("#64748b") // Faint gray

	// Special effects
	glow  = lipgloss.Color("#3b82f6") // Blue glow
	pulse = lipgloss.Color("#ec4899") // Pink pulse
)

var (
	quitKeys   = key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("", "quit"))
	killKeys   = key.NewBinding(key.WithKeys("k"), key.WithHelp("", "kill process"))
	searchKeys = key.NewBinding(key.WithKeys("/"), key.WithHelp("", "search"))
	tabKeys    = key.NewBinding(key.WithKeys("tab"), key.WithHelp("", "switch view"))
)

func initialModel() model {
	numCores := runtime.NumCPU()
	m := model{
		viewMode:       "overview",
		orbs:           make([]orb, 8),
		lightRays:      make([]lightRay, 12),
		matrixRain:     make([]matrixDrop, 20),
		energyField:    make([]energyNode, 15),
		cpuHistory:     make([][]float64, numCores),
		memHistory:     make([]float64, 60),
		bandwidthIn:    make([]float64, 60),
		bandwidthOut:   make([]float64, 60),
		networkTraffic: make([]float64, 60),
		sortColumn:     "CPU",
		sortReverse:    true,
		selectedRow:    0,
		scrollOffset:   0,
		selectedPanel:  0,
		glowIntensity:  0.5,
		pulsePhase:     0,
	}

	// Initialize CPU history
	for i := range m.cpuHistory {
		m.cpuHistory[i] = make([]float64, 60)
	}

	// Initialize elegant floating orbs
	colors := []lipgloss.Color{primary, secondary, accent, success, warning}
	for i := range m.orbs {
		m.orbs[i] = orb{
			x:           rand.Float64() * 120,
			y:           rand.Float64() * 40,
			vx:          (rand.Float64() - 0.5) * 0.2,
			vy:          (rand.Float64() - 0.5) * 0.15,
			radius:      rand.Float64()*2 + 1,
			intensity:   rand.Float64()*0.8 + 0.2,
			color:       colors[rand.Intn(len(colors))],
			pulsePeriod: rand.Float64()*4 + 2,
			pulseOffset: rand.Float64() * 6.28,
		}
	}

	// Initialize light rays
	for i := range m.lightRays {
		m.lightRays[i] = lightRay{
			x:         rand.Float64() * 120,
			y:         rand.Float64() * 40,
			angle:     rand.Float64() * 6.28,
			length:    rand.Float64()*15 + 5,
			intensity: rand.Float64()*0.6 + 0.1,
			color:     colors[rand.Intn(len(colors))],
		}
	}

	// Initialize matrix rain
	chars := []rune("ﾊﾐﾋｰｳｼﾅﾓﾆｻﾜﾂｵﾘｱﾎﾃﾏｹﾒｴｶｷﾑﾕﾗｾﾈｽﾀﾇﾍ0123456789")
	for i := range m.matrixRain {
		m.matrixRain[i] = matrixDrop{
			x:         float64(rand.Intn(120)),
			y:         rand.Float64() * 40,
			speed:     rand.Float64()*0.3 + 0.1,
			length:    rand.Intn(8) + 3,
			chars:     make([]rune, rand.Intn(8)+3),
			intensity: rand.Float64()*0.7 + 0.3,
		}
		for j := range m.matrixRain[i].chars {
			m.matrixRain[i].chars[j] = chars[rand.Intn(len(chars))]
		}
	}

	// Initialize energy field
	for i := range m.energyField {
		connections := make([]int, 0)
		// Connect to 1-3 random other nodes
		numConnections := rand.Intn(3) + 1
		for j := 0; j < numConnections; j++ {
			target := rand.Intn(len(m.energyField))
			if target != i {
				connections = append(connections, target)
			}
		}

		m.energyField[i] = energyNode{
			x:           rand.Float64() * 120,
			y:           rand.Float64() * 40,
			connections: connections,
			pulsePhase:  rand.Float64() * 6.28,
			energy:      rand.Float64()*0.8 + 0.2,
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
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
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
	interfaces  []NetworkInterface
	lastUpdate  time.Time
}

type alertData struct {
	alerts     []SystemAlert
	lastUpdate time.Time
}

func fetchNetworkData() tea.Cmd {
	return func() tea.Msg {
		connections := getNetworkConnections()
		interfaces := getNetworkInterfaces()
		return networkData{
			connections: connections,
			interfaces:  interfaces,
			lastUpdate:  time.Now(),
		}
	}
}

func fetchSystemAlerts() tea.Cmd {
	return func() tea.Msg {
		alerts := getSystemAlerts()
		return alertData{
			alerts:     alerts,
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
				visibleRows := m.height - 15
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
		m.pulsePhase += 0.05
		m.glowIntensity = 0.3 + 0.2*math.Sin(m.pulsePhase)

		// Update elegant orbs
		for i := range m.orbs {
			orb := &m.orbs[i]
			orb.x += orb.vx
			orb.y += orb.vy

			// Gentle bouncing with smooth curves
			if orb.x <= 0 || orb.x >= 120 {
				orb.vx *= -0.8
				orb.x = math.Max(0, math.Min(120, orb.x))
			}
			if orb.y <= 0 || orb.y >= 40 {
				orb.vy *= -0.8
				orb.y = math.Max(0, math.Min(40, orb.y))
			}

			// Update pulse
			orb.intensity = 0.3 + 0.5*math.Sin(float64(m.frame)*0.02*orb.pulsePeriod+orb.pulseOffset)
		}

		// Update light rays
		for i := range m.lightRays {
			ray := &m.lightRays[i]
			ray.angle += 0.01
			ray.intensity = 0.2 + 0.4*math.Sin(float64(m.frame)*0.03+float64(i))
		}

		// Update matrix rain
		for i := range m.matrixRain {
			drop := &m.matrixRain[i]
			drop.y += drop.speed
			if drop.y > 40 {
				drop.y = -float64(drop.length)
				drop.x = float64(rand.Intn(120))
			}
		}

		// Update energy field
		for i := range m.energyField {
			node := &m.energyField[i]
			node.pulsePhase += 0.02
			node.energy = 0.4 + 0.4*math.Sin(node.pulsePhase)
		}

		// Fetch new data periodically
		var cmds []tea.Cmd
		if m.frame%40 == 0 { // 2 seconds
			cmds = append(cmds, fetchProcesses())
		}
		if m.frame%50 == 0 { // 2.5 seconds
			cmds = append(cmds, fetchNetworkData())
		}
		if m.frame%80 == 0 { // 4 seconds
			cmds = append(cmds, fetchSystemAlerts())
		}

		cmds = append(cmds, tick())
		return m, tea.Batch(cmds...)

	case processData:
		m.processes = msg.processes
		m.systemStats = msg.systemStats

		// Update histories smoothly
		for i, usage := range msg.systemStats.CPUUsage {
			if i < len(m.cpuHistory) {
				m.cpuHistory[i] = append(m.cpuHistory[i][1:], usage)
			}
		}

		memUsagePercent := float64(msg.systemStats.MemUsed) / float64(msg.systemStats.MemTotal) * 100
		m.memHistory = append(m.memHistory[1:], memUsagePercent)

		m.sortProcesses()
		m.filteredProcs = filterProcesses(m.processes, m.searchQuery)

		if m.selectedRow >= len(m.filteredProcs) {
			m.selectedRow = max(0, len(m.filteredProcs)-1)
		}

		return m, nil

	case networkData:
		m.connections = msg.connections
		m.interfaces = msg.interfaces

		// Update network traffic history
		totalTraffic := 0.0
		for _, iface := range m.interfaces {
			totalTraffic += float64(iface.BytesIn+iface.BytesOut) / 1024 / 1024 // MB
		}
		m.networkTraffic = append(m.networkTraffic[1:], totalTraffic)

		return m, nil

	case alertData:
		m.alerts = msg.alerts
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().
			Foreground(primary).
			Align(lipgloss.Center).
			Render("Disconnecting from the matrix...\n")
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing elegant interface..."
	}

	// Create main layout
	header := m.renderHeader()
	tabs := m.renderViewTabs()
	content := m.renderContent()
	controls := m.renderControls()

	// Add subtle background effects
	bg := m.renderBackgroundEffects()

	layout := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabs,
		content,
		controls,
	)

	// Overlay background effects
	return m.overlayEffects(layout, bg)
}

func (m model) renderHeader() string {
	title := "⟨ QUANTUM COMMAND CENTER ⟩"

	titleStyle := lipgloss.NewStyle().
		Foreground(primary).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width)

	// Add subtle glow effect
	glowStyle := lipgloss.NewStyle().
		Foreground(glow).
		Faint(true).
		Align(lipgloss.Center).
		Width(m.width)

	time := lipgloss.NewStyle().
		Foreground(textMuted).
		Align(lipgloss.Right).
		Width(m.width).
		Render(fmt.Sprintf("◉ %s", time.Now().Format("15:04:05")))

	return lipgloss.JoinVertical(lipgloss.Left,
		glowStyle.Render("═══════════════════════════════════════"),
		titleStyle.Render(title),
		glowStyle.Render("═══════════════════════════════════════"),
		time,
		"",
	)
}

func (m model) renderViewTabs() string {
	tabs := []struct {
		key   string
		name  string
		icon  string
		color lipgloss.Color
	}{
		{"1", "Overview", "⬢", primary},
		{"2", "Processes", "⚡", secondary},
		{"3", "Network", "◈", accent},
		{"4", "Alerts", "⚠", warning},
	}

	activeIndex := 0
	switch m.viewMode {
	case "overview":
		activeIndex = 0
	case "processes":
		activeIndex = 1
	case "network":
		activeIndex = 2
	case "alerts":
		activeIndex = 3
	}

	var tabStrs []string
	for i, tab := range tabs {
		content := fmt.Sprintf("%s %s %s", tab.icon, tab.key, tab.name)

		if i == activeIndex {
			// Active tab with elegant styling
			tabStrs = append(tabStrs, lipgloss.NewStyle().
				Foreground(bg).
				Background(tab.color).
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(tab.color).
				Render(content))
		} else {
			// Inactive tab with subtle styling
			tabStrs = append(tabStrs, lipgloss.NewStyle().
				Foreground(tab.color).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(textFaint).
				Padding(0, 2).
				Render(content))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, tabStrs...) + "\n"
}

func (m model) renderContent() string {
	switch m.viewMode {
	case "overview":
		return m.renderOverview()
	case "processes":
		return m.renderProcesses()
	case "network":
		return m.renderNetwork()
	case "alerts":
		return m.renderAlerts()
	default:
		return m.renderOverview()
	}
}

func (m model) renderOverview() string {
	leftPanel := m.renderSystemPanel()
	centerPanel := m.renderNetworkPanel()
	rightPanel := m.renderAlertsPanel()

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Height(m.height - 12)

	leftStyled := panelStyle.Copy().
		BorderForeground(primary).
		Width((m.width / 3) - 2).
		Render(leftPanel)

	centerStyled := panelStyle.Copy().
		BorderForeground(accent).
		Width((m.width / 3) - 2).
		Render(centerPanel)

	rightStyled := panelStyle.Copy().
		BorderForeground(warning).
		Width((m.width / 3) - 2).
		Render(rightPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, centerStyled, rightStyled)
}

func (m model) renderSystemPanel() string {
	var s strings.Builder

	// Title with icon
	title := lipgloss.NewStyle().
		Foreground(primary).
		Bold(true).
		Render("⬢ SYSTEM STATUS")
	s.WriteString(title + "\n\n")

	// CPU information
	if len(m.systemStats.CPUUsage) > 0 {
		avgCPU := 0.0
		for _, usage := range m.systemStats.CPUUsage {
			avgCPU += usage
		}
		avgCPU /= float64(len(m.systemStats.CPUUsage))

		cpuLabel := lipgloss.NewStyle().Foreground(text).Render("CPU Usage:")
		cpuBar := m.renderElegantBar(avgCPU, 100, 20, success, danger)
		cpuText := lipgloss.NewStyle().Foreground(textMuted).Render(fmt.Sprintf("%.1f%%", avgCPU))

		s.WriteString(fmt.Sprintf("%s\n%s %s\n\n", cpuLabel, cpuBar, cpuText))

		// CPU cores
		s.WriteString(lipgloss.NewStyle().Foreground(textMuted).Render("Cores:\n"))
		for i, usage := range m.systemStats.CPUUsage {
			if i >= 8 {
				break
			} // Limit display
			coreBar := m.renderMiniBar(usage, 100, 12, success, danger)
			s.WriteString(fmt.Sprintf("C%d %s %.1f%%\n", i, coreBar, usage))
		}
		s.WriteString("\n")
	}

	// Memory information
	if m.systemStats.MemTotal > 0 {
		memUsage := float64(m.systemStats.MemUsed) / float64(m.systemStats.MemTotal) * 100

		memLabel := lipgloss.NewStyle().Foreground(text).Render("Memory Usage:")
		memBar := m.renderElegantBar(memUsage, 100, 20, success, danger)
		memText := lipgloss.NewStyle().Foreground(textMuted).Render(fmt.Sprintf("%.1f%% (%s/%s)",
			memUsage, formatBytes(m.systemStats.MemUsed), formatBytes(m.systemStats.MemTotal)))

		s.WriteString(fmt.Sprintf("%s\n%s\n%s\n\n", memLabel, memBar, memText))
	}

	// Load average
	loadText := lipgloss.NewStyle().Foreground(text).Render("Load Average:")
	loadValues := lipgloss.NewStyle().Foreground(textMuted).Render(
		fmt.Sprintf("%.2f %.2f %.2f", m.systemStats.LoadAvg[0], m.systemStats.LoadAvg[1], m.systemStats.LoadAvg[2]))
	s.WriteString(fmt.Sprintf("%s\n%s\n\n", loadText, loadValues))

	// Uptime
	uptimeText := lipgloss.NewStyle().Foreground(text).Render("Uptime:")
	uptimeValue := lipgloss.NewStyle().Foreground(textMuted).Render(formatDuration(m.systemStats.Uptime))
	s.WriteString(fmt.Sprintf("%s\n%s", uptimeText, uptimeValue))

	return s.String()
}

func (m model) renderNetworkPanel() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Render("◈ NETWORK STATUS")
	s.WriteString(title + "\n\n")

	// Network interfaces
	s.WriteString(lipgloss.NewStyle().Foreground(text).Render("Interfaces:\n"))
	for _, iface := range m.interfaces {
		if len(iface.Name) == 0 {
			continue
		}

		ifaceStyle := lipgloss.NewStyle().
			Foreground(textMuted).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(accent).
			PaddingLeft(1)

		content := fmt.Sprintf("%s\n↑ %s  ↓ %s",
			iface.Name, formatBytes(iface.BytesOut), formatBytes(iface.BytesIn))

		s.WriteString(ifaceStyle.Render(content) + "\n\n")
	}

	// Active connections
	s.WriteString(lipgloss.NewStyle().Foreground(text).Render("Active Connections:\n"))
	connectionCount := min(len(m.connections), 8)
	for i := 0; i < connectionCount; i++ {
		conn := m.connections[i]

		stateColor := success
		switch conn.State {
		case "ESTABLISHED":
			stateColor = success
		case "LISTEN":
			stateColor = accent
		case "TIME_WAIT":
			stateColor = warning
		default:
			stateColor = textMuted
		}

		connText := fmt.Sprintf("%-4s %s:%s → %s",
			conn.Protocol, conn.LocalAddr, conn.LocalPort, conn.State)
		if len(connText) > 28 {
			connText = connText[:25] + "..."
		}

		s.WriteString(lipgloss.NewStyle().Foreground(stateColor).Render(connText) + "\n")
	}

	// Network activity visualization
	s.WriteString("\n" + lipgloss.NewStyle().Foreground(text).Render("Activity:") + "\n")
	activityViz := m.renderNetworkActivity()
	s.WriteString(activityViz)

	return s.String()
}

func (m model) renderAlertsPanel() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Foreground(warning).
		Bold(true).
		Render("⚠ SYSTEM ALERTS")
	s.WriteString(title + "\n\n")

	if len(m.alerts) == 0 {
		noAlerts := lipgloss.NewStyle().
			Foreground(success).
			Render("✓ All systems nominal")
		s.WriteString(noAlerts)
		return s.String()
	}

	// Show recent alerts with elegant styling
	maxAlerts := min(len(m.alerts), 10)
	for i := 0; i < maxAlerts; i++ {
		alert := m.alerts[i]

		var levelColor lipgloss.Color
		var icon string
		switch alert.Level {
		case "ERROR":
			levelColor = danger
			icon = "●"
		case "WARN":
			levelColor = warning
			icon = "◐"
		case "INFO":
			levelColor = accent
			icon = "◯"
		default:
			levelColor = textMuted
			icon = "○"
		}

		timestamp := alert.Timestamp.Format("15:04")
		message := alert.Message
		if len(message) > 30 {
			message = message[:27] + "..."
		}

		alertStyle := lipgloss.NewStyle().
			Foreground(levelColor).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(levelColor).
			PaddingLeft(1)

		content := fmt.Sprintf("%s %s\n%s", icon, timestamp, message)
		s.WriteString(alertStyle.Render(content) + "\n\n")
	}

	return s.String()
}

func (m model) renderProcesses() string {
	var s strings.Builder

	// Header
	title := fmt.Sprintf("⚡ PROCESS MONITOR (%d processes)", len(m.filteredProcs))
	if m.searchQuery != "" {
		title += fmt.Sprintf(" - Filtered: '%s'", m.searchQuery)
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(secondary).
		Bold(true)
	s.WriteString(headerStyle.Render(title) + "\n\n")

	// Column headers with elegant styling
	headers := []string{"PID", "USER", "CPU%", "MEM%", "STATE", "COMMAND"}
	widths := []int{8, 12, 8, 8, 8, m.width - 50}

	var headerParts []string
	for i, header := range headers {
		headerStyle := lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Width(widths[i]).
			Align(lipgloss.Left)
		headerParts = append(headerParts, headerStyle.Render(header))
	}

	s.WriteString(strings.Join(headerParts, " ") + "\n")
	s.WriteString(lipgloss.NewStyle().
		Foreground(textFaint).
		Render(strings.Repeat("─", m.width-2)) + "\n\n")

	// Process list
	visibleRows := m.height - 15
	startIdx := m.scrollOffset
	endIdx := min(startIdx+visibleRows, len(m.filteredProcs))

	for i := startIdx; i < endIdx; i++ {
		if i >= len(m.filteredProcs) {
			break
		}

		proc := m.filteredProcs[i]
		isSelected := i == m.selectedRow

		// Dynamic colors based on values
		var cpuColor, memColor, stateColor lipgloss.Color

		if proc.CPU > 80 {
			cpuColor = danger
		} else if proc.CPU > 40 {
			cpuColor = warning
		} else {
			cpuColor = success
		}

		if proc.Memory > 10 {
			memColor = danger
		} else if proc.Memory > 5 {
			memColor = warning
		} else {
			memColor = success
		}

		switch proc.State {
		case "R", "R+":
			stateColor = success
		case "S", "S+":
			stateColor = accent
		case "T":
			stateColor = warning
		case "Z":
			stateColor = danger
		default:
			stateColor = textMuted
		}

		// Format fields
		pidText := fmt.Sprintf("%d", proc.PID)
		userText := proc.User
		cpuText := fmt.Sprintf("%.1f", proc.CPU)
		memText := fmt.Sprintf("%.1f", proc.Memory)
		stateText := proc.State

		command := proc.Command
		if len(command) > widths[5] {
			command = command[:widths[5]-3] + "..."
		}

		// Create row parts
		parts := []string{
			lipgloss.NewStyle().Width(widths[0]).Render(pidText),
			lipgloss.NewStyle().Width(widths[1]).Render(userText),
			lipgloss.NewStyle().Width(widths[2]).Foreground(cpuColor).Render(cpuText),
			lipgloss.NewStyle().Width(widths[3]).Foreground(memColor).Render(memText),
			lipgloss.NewStyle().Width(widths[4]).Foreground(stateColor).Render(stateText),
			lipgloss.NewStyle().Foreground(text).Render(command),
		}

		row := strings.Join(parts, " ")

		if isSelected {
			row = lipgloss.NewStyle().
				Foreground(bg).
				Background(primary).
				Bold(true).
				Width(m.width - 2).
				Render(" " + row + " ")
		}

		s.WriteString(row + "\n")
	}

	return s.String()
}

func (m model) renderNetwork() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Render("◈ NETWORK MONITOR")
	s.WriteString(title + "\n\n")

	// Interfaces section
	s.WriteString(lipgloss.NewStyle().Foreground(text).Bold(true).Render("Network Interfaces:") + "\n")
	for _, iface := range m.interfaces {
		if len(iface.Name) == 0 {
			continue
		}

		ifaceBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1).
			Margin(0, 1, 1, 0).
			Width(30)

		content := fmt.Sprintf("%s\n\n↑ Out: %s\n↓ In:  %s\n📦 Packets: %d/%d",
			lipgloss.NewStyle().Foreground(accent).Bold(true).Render(iface.Name),
			formatBytes(iface.BytesOut),
			formatBytes(iface.BytesIn),
			iface.PacketsOut, iface.PacketsIn)

		s.WriteString(ifaceBox.Render(content) + "\n")
	}

	// Connections section
	s.WriteString(lipgloss.NewStyle().Foreground(text).Bold(true).Render("Active Connections:") + "\n\n")

	// Connection headers
	connHeaders := []string{"PROTO", "LOCAL", "REMOTE", "STATE", "PROCESS"}
	connWidths := []int{8, 22, 22, 12, m.width - 70}

	var headerParts []string
	for i, header := range connHeaders {
		headerStyle := lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Width(connWidths[i])
		headerParts = append(headerParts, headerStyle.Render(header))
	}

	s.WriteString(strings.Join(headerParts, " ") + "\n")
	s.WriteString(strings.Repeat("─", m.width-2) + "\n")

	// Connection list
	maxConns := min(len(m.connections), m.height-18)
	for i := 0; i < maxConns; i++ {
		conn := m.connections[i]

		stateColor := success
		switch conn.State {
		case "ESTABLISHED":
			stateColor = success
		case "LISTEN":
			stateColor = accent
		case "TIME_WAIT":
			stateColor = warning
		default:
			stateColor = textMuted
		}

		local := fmt.Sprintf("%s:%s", conn.LocalAddr, conn.LocalPort)
		remote := fmt.Sprintf("%s:%s", conn.RemoteAddr, conn.RemotePort)

		if len(local) > connWidths[1] {
			local = local[:connWidths[1]-3] + "..."
		}
		if len(remote) > connWidths[2] {
			remote = remote[:connWidths[2]-3] + "..."
		}
		if len(conn.ProcessName) > connWidths[4] {
			conn.ProcessName = conn.ProcessName[:connWidths[4]-3] + "..."
		}

		parts := []string{
			lipgloss.NewStyle().Width(connWidths[0]).Render(conn.Protocol),
			lipgloss.NewStyle().Width(connWidths[1]).Render(local),
			lipgloss.NewStyle().Width(connWidths[2]).Render(remote),
			lipgloss.NewStyle().Width(connWidths[3]).Foreground(stateColor).Render(conn.State),
			lipgloss.NewStyle().Foreground(textMuted).Render(conn.ProcessName),
		}

		s.WriteString(strings.Join(parts, " ") + "\n")
	}

	return s.String()
}

func (m model) renderAlerts() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Foreground(warning).
		Bold(true).
		Render("⚠ SYSTEM ALERTS")
	s.WriteString(title + "\n\n")

	if len(m.alerts) == 0 {
		noAlerts := lipgloss.NewStyle().
			Foreground(success).
			Align(lipgloss.Center).
			Width(m.width).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(success).
			Padding(2).
			Render("✓ No Active Alerts\n\nAll systems are operating normally")

		s.WriteString(noAlerts)
		return s.String()
	}

	// Alert list with elegant cards
	maxAlerts := min(len(m.alerts), m.height-10)
	for i := 0; i < maxAlerts; i++ {
		alert := m.alerts[i]

		var levelColor lipgloss.Color
		var icon string
		var severity string

		switch alert.Level {
		case "ERROR":
			levelColor = danger
			icon = "🔴"
			severity = "CRITICAL"
		case "WARN":
			levelColor = warning
			icon = "🟡"
			severity = "WARNING"
		case "INFO":
			levelColor = accent
			icon = "🔵"
			severity = "INFO"
		default:
			levelColor = textMuted
			icon = "⚪"
			severity = "NOTICE"
		}

		timestamp := alert.Timestamp.Format("2006-01-02 15:04:05")

		alertCard := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(levelColor).
			Padding(1).
			Margin(0, 0, 1, 0).
			Width(m.width - 4)

		header := lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(levelColor).Bold(true).Render(fmt.Sprintf("%s %s", icon, severity)),
			lipgloss.NewStyle().Foreground(textMuted).Align(lipgloss.Right).Width(m.width-20).Render(timestamp),
		)

		content := fmt.Sprintf("%s\n\n%s\n\nSource: %s • Type: %s",
			header,
			lipgloss.NewStyle().Foreground(text).Render(alert.Message),
			lipgloss.NewStyle().Foreground(textMuted).Render(alert.Source),
			lipgloss.NewStyle().Foreground(textMuted).Render(string(alert.Type)))

		s.WriteString(alertCard.Render(content) + "\n")
	}

	return s.String()
}

func (m model) renderControls() string {
	var controls []string

	switch m.viewMode {
	case "processes":
		if m.searchMode {
			controls = append(controls,
				lipgloss.NewStyle().Foreground(warning).Render("🔍 SEARCH: "+m.searchQuery+"_"),
				lipgloss.NewStyle().Foreground(textMuted).Render("ESC: cancel"),
			)
		} else {
			controls = append(controls,
				lipgloss.NewStyle().Foreground(primary).Render("↑↓: navigate"),
				lipgloss.NewStyle().Foreground(danger).Render("k: kill"),
				lipgloss.NewStyle().Foreground(warning).Render("/: search"),
				lipgloss.NewStyle().Foreground(accent).Render("c/m/p: sort"),
			)
		}
	default:
		controls = append(controls,
			lipgloss.NewStyle().Foreground(primary).Render("TAB: cycle views"),
			lipgloss.NewStyle().Foreground(accent).Render("1-4: quick switch"),
		)
	}

	controls = append(controls, lipgloss.NewStyle().Foreground(danger).Render("q: quit"))

	return lipgloss.NewStyle().
		Background(bgLight).
		Foreground(text).
		Padding(1, 2).
		Width(m.width).
		Align(lipgloss.Center).
		Render(strings.Join(controls, "  •  "))
}

// Enhanced visual effects
func (m model) renderElegantBar(value, max float64, width int, lowColor, highColor lipgloss.Color) string {
	if max == 0 {
		return ""
	}

	ratio := value / max
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}

	// Create gradient effect
	var segments []string
	for i := 0; i < width; i++ {
		if i < filled {
			if ratio > 0.7 {
				segments = append(segments, lipgloss.NewStyle().Foreground(highColor).Render("█"))
			} else if ratio > 0.4 {
				segments = append(segments, lipgloss.NewStyle().Foreground(warning).Render("█"))
			} else {
				segments = append(segments, lipgloss.NewStyle().Foreground(lowColor).Render("█"))
			}
		} else {
			segments = append(segments, lipgloss.NewStyle().Foreground(textFaint).Render("░"))
		}
	}

	return strings.Join(segments, "")
}

func (m model) renderMiniBar(value, max float64, width int, lowColor, highColor lipgloss.Color) string {
	if max == 0 {
		return ""
	}

	ratio := value / max
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)

	color := lowColor
	if ratio > 0.7 {
		color = highColor
	} else if ratio > 0.4 {
		color = warning
	}

	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

func (m model) renderNetworkActivity() string {
	var lines []string

	// Create a wave pattern for network activity
	for row := 0; row < 6; row++ {
		var line strings.Builder
		for col := 0; col < 25; col++ {
			char := "·"

			// Create flowing wave pattern
			waveHeight := 2.0 + 1.5*math.Sin(float64(col)*0.3+float64(m.frame)*0.05)
			if math.Abs(float64(row)-waveHeight) < 0.8 {
				char = "~"
			}

			// Add data flow indicators
			if (col+m.frame/4)%8 == 0 {
				char = "◦"
			}

			line.WriteString(char)
		}

		color := accent
		if row == 2 || row == 3 {
			color = primary
		}

		lines = append(lines, lipgloss.NewStyle().Foreground(color).Render(line.String()))
	}

	return strings.Join(lines, "\n")
}

func (m model) renderBackgroundEffects() string {
	// This would render subtle background effects
	// For now, return empty to keep focus on content
	return ""
}

func (m model) overlayEffects(content, effects string) string {
	// Simple overlay - in a full implementation, this would blend effects with content
	return content
}

// Utility functions
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

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// min function is defined in system_monitor.go

// max function is defined in system_monitor.go

// Process management functions
func (m *model) sortProcesses() {
	if len(m.processes) == 0 {
		return
	}

	sort.Slice(m.processes, func(i, j int) bool {
		var less bool
		switch m.sortColumn {
		case "CPU":
			less = m.processes[i].CPU < m.processes[j].CPU
		case "Memory":
			less = m.processes[i].Memory < m.processes[j].Memory
		case "PID":
			less = m.processes[i].PID < m.processes[j].PID
		default:
			less = m.processes[i].CPU < m.processes[j].CPU
		}
		if m.sortReverse {
			less = !less
		}
		return less
	})
}

func filterProcesses(processes []Process, query string) []Process {
	if query == "" {
		return processes
	}

	query = strings.ToLower(query)
	var filtered []Process

	for _, proc := range processes {
		if strings.Contains(strings.ToLower(proc.Name), query) ||
			strings.Contains(strings.ToLower(proc.Command), query) ||
			strings.Contains(strings.ToLower(proc.User), query) ||
			strings.Contains(fmt.Sprintf("%d", proc.PID), query) {
			filtered = append(filtered, proc)
		}
	}

	return filtered
}

func killProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

// Data fetching functions (enhanced implementations)
func getProcessesAndStats() ([]Process, SystemStats) {
	var processes []Process
	stats := SystemStats{
		CPUCores: runtime.NumCPU(),
		CPUUsage: make([]float64, runtime.NumCPU()),
	}

	// Get CPU usage per core
	for i := 0; i < runtime.NumCPU(); i++ {
		stats.CPUUsage[i] = rand.Float64() * 100
	}

	// Simulate memory stats
	stats.MemTotal = 8 * 1024 * 1024 * 1024 // 8GB
	stats.MemUsed = uint64(rand.Float64() * float64(stats.MemTotal) * 0.8)
	stats.LoadAvg = [3]float64{rand.Float64() * 2, rand.Float64() * 2, rand.Float64() * 2}
	stats.Uptime = time.Hour * 24 * time.Duration(rand.Intn(30))

	// Get process list via ps
	cmd := exec.Command("ps", "axo", "pid,user,pcpu,pmem,state,ppid,etime,command")
	output, err := cmd.Output()
	if err != nil {
		return processes, stats
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		ppid, _ := strconv.Atoi(fields[5])

		command := strings.Join(fields[7:], " ")
		name := fields[7]
		if strings.Contains(name, "/") {
			parts := strings.Split(name, "/")
			name = parts[len(parts)-1]
		}

		processes = append(processes, Process{
			PID:       pid,
			Name:      name,
			User:      fields[1],
			CPU:       cpu,
			Memory:    mem,
			State:     fields[4],
			PPID:      ppid,
			StartTime: fields[6],
			Command:   command,
		})
	}

	return processes, stats
}

func getNetworkConnections() []NetworkConnection {
	var connections []NetworkConnection

	// Simulate some connections
	protocols := []string{"TCP", "UDP"}
	states := []string{"ESTABLISHED", "LISTEN", "TIME_WAIT", "CLOSE_WAIT"}

	for i := 0; i < 15; i++ {
		conn := NetworkConnection{
			Protocol:    protocols[rand.Intn(len(protocols))],
			LocalAddr:   fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			LocalPort:   fmt.Sprintf("%d", rand.Intn(65535)+1),
			RemoteAddr:  fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)),
			RemotePort:  fmt.Sprintf("%d", rand.Intn(65535)+1),
			State:       states[rand.Intn(len(states))],
			PID:         rand.Intn(30000) + 1000,
			ProcessName: []string{"chrome", "ssh", "node", "python", "go", "nginx"}[rand.Intn(6)],
		}
		connections = append(connections, conn)
	}

	return connections
}

func getNetworkInterfaces() []NetworkInterface {
	interfaces := []NetworkInterface{
		{
			Name:       "en0",
			BytesIn:    uint64(rand.Intn(1000000000)),
			BytesOut:   uint64(rand.Intn(1000000000)),
			PacketsIn:  uint64(rand.Intn(1000000)),
			PacketsOut: uint64(rand.Intn(1000000)),
			Speed:      1000000000, // 1 Gbps
		},
		{
			Name:       "lo0",
			BytesIn:    uint64(rand.Intn(100000000)),
			BytesOut:   uint64(rand.Intn(100000000)),
			PacketsIn:  uint64(rand.Intn(100000)),
			PacketsOut: uint64(rand.Intn(100000)),
			Speed:      0,
		},
	}

	return interfaces
}

func getSystemAlerts() []SystemAlert {
	var alerts []SystemAlert

	// Simulate some alerts
	alertTypes := []AlertType{AlertTypeCPU, AlertTypeMemory, AlertTypeDisk, AlertTypeNetwork, AlertTypeProcess}
	levels := []AlertLevel{AlertLevelInfo, AlertLevelWarning, AlertLevelError}
	messages := []string{
		"High CPU usage detected",
		"Memory usage above threshold",
		"Disk space running low",
		"Network connection timeout",
		"Process crashed unexpectedly",
		"Temperature warning",
		"Load average high",
	}

	for i := 0; i < rand.Intn(5); i++ {
		alert := SystemAlert{
			Type:      alertTypes[rand.Intn(len(alertTypes))],
			Level:     levels[rand.Intn(len(levels))],
			Message:   messages[rand.Intn(len(messages))],
			Timestamp: time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second),
			Source:    "system-monitor",
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
