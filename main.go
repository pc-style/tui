package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type visualMode int

const (
	modeMatrix visualMode = iota
	modeFire
	modeWaves
	modeParticles
	modeDrawing
)

type tickMsg struct{}

type model struct {
	mode     visualMode
	width    int
	height   int
	frame    int
	quitting bool
	err      error
	
	// Matrix effect
	matrixChars [][]rune
	matrixColors [][]int
	matrixSpeed []int
	
	// Fire effect
	fireBuffer [][]int
	
	// Wave effect
	wavePhase float64
	
	// Particle system
	particles []particle
	
	// Drawing mode
	canvas [][]rune
}

type particle struct {
	x, y   float64
	vx, vy float64
	life   int
	char   rune
	color  int
}

var (
	quitKeys = key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("", "press q to quit"),
	)
	
	modeKeys = key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5"),
		key.WithHelp("", "1-5 to switch modes"),
	)
	
	matrixChars = []rune("アカサタナハマヤラワガザダバパイキシチニヒミイリヰギジヂビピウクスツヌフムユルグズブプエケセテネヘメエレヱゲゼデベペオコソトノホモヨロヲゴゾドボポ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func initialModel() model {
	return model{
		mode:        modeMatrix,
		frame:       0,
		matrixChars: [][]rune{},
		matrixColors: [][]int{},
		matrixSpeed: []int{},
		fireBuffer:  [][]int{},
		wavePhase:   0,
		particles:   []particle{},
		canvas:      [][]rune{},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg { return tickMsg{} }),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.initializeMode()
		
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		
		switch msg.String() {
		case "1":
			m.mode = modeMatrix
			m = m.initializeMode()
		case "2":
			m.mode = modeFire
			m = m.initializeMode()
		case "3":
			m.mode = modeWaves
			m = m.initializeMode()
		case "4":
			m.mode = modeParticles
			m = m.initializeMode()
		case "5":
			m.mode = modeDrawing
			m = m.initializeMode()
		}
		
	case tickMsg:
		m.frame++
		m = m.updateAnimation()
		return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg { return tickMsg{} })
		
	case errMsg:
		m.err = msg
		return m, nil
	}
	
	return m, nil
}

func (m model) initializeMode() model {
	if m.width == 0 || m.height == 0 {
		return m
	}
	
	switch m.mode {
	case modeMatrix:
		m.matrixChars = make([][]rune, m.width)
		m.matrixColors = make([][]int, m.width)
		m.matrixSpeed = make([]int, m.width)
		for i := range m.matrixChars {
			m.matrixChars[i] = make([]rune, m.height)
			m.matrixColors[i] = make([]int, m.height)
			m.matrixSpeed[i] = rand.Intn(3) + 1
			for j := range m.matrixChars[i] {
				m.matrixChars[i][j] = matrixChars[rand.Intn(len(matrixChars))]
				m.matrixColors[i][j] = rand.Intn(m.height)
			}
		}
		
	case modeFire:
		m.fireBuffer = make([][]int, m.width)
		for i := range m.fireBuffer {
			m.fireBuffer[i] = make([]int, m.height)
		}
		
	case modeParticles:
		m.particles = make([]particle, 200)
		for i := range m.particles {
			m.particles[i] = particle{
				x:     rand.Float64() * float64(m.width),
				y:     rand.Float64() * float64(m.height),
				vx:    (rand.Float64() - 0.5) * 2,
				vy:    (rand.Float64() - 0.5) * 2,
				life:  rand.Intn(100) + 50,
				char:  matrixChars[rand.Intn(len(matrixChars))],
				color: rand.Intn(255),
			}
		}
		
	case modeDrawing:
		m.canvas = make([][]rune, m.width)
		for i := range m.canvas {
			m.canvas[i] = make([]rune, m.height)
			for j := range m.canvas[i] {
				m.canvas[i][j] = ' '
			}
		}
	}
	
	return m
}

func (m model) updateAnimation() model {
	switch m.mode {
	case modeMatrix:
		for i := range m.matrixChars {
			if m.frame%m.matrixSpeed[i] == 0 {
				for j := m.height - 1; j > 0; j-- {
					m.matrixChars[i][j] = m.matrixChars[i][j-1]
					m.matrixColors[i][j] = m.matrixColors[i][j-1]
				}
				m.matrixChars[i][0] = matrixChars[rand.Intn(len(matrixChars))]
				m.matrixColors[i][0] = 0
				
				for j := range m.matrixColors[i] {
					if m.matrixColors[i][j] < m.height-1 {
						m.matrixColors[i][j]++
					}
				}
			}
		}
		
	case modeFire:
		// Create fire at bottom
		for i := 0; i < m.width; i++ {
			m.fireBuffer[i][m.height-1] = 255
		}
		
		// Propagate fire upward
		for y := m.height - 2; y >= 0; y-- {
			for x := 0; x < m.width; x++ {
				decay := rand.Intn(3)
				srcIdx := x + rand.Intn(3) - 1
				if srcIdx >= 0 && srcIdx < m.width {
					newVal := m.fireBuffer[srcIdx][y+1] - decay
					if newVal < 0 {
						newVal = 0
					}
					m.fireBuffer[x][y] = newVal
				}
			}
		}
		
	case modeWaves:
		m.wavePhase += 0.1
		
	case modeParticles:
		for i := range m.particles {
			p := &m.particles[i]
			p.x += p.vx
			p.y += p.vy
			p.life--
			
			// Bounce off edges
			if p.x <= 0 || p.x >= float64(m.width-1) {
				p.vx = -p.vx
			}
			if p.y <= 0 || p.y >= float64(m.height-1) {
				p.vy = -p.vy
			}
			
			// Reset particle if life expired
			if p.life <= 0 {
				p.x = rand.Float64() * float64(m.width)
				p.y = rand.Float64() * float64(m.height)
				p.vx = (rand.Float64() - 0.5) * 2
				p.vy = (rand.Float64() - 0.5) * 2
				p.life = rand.Intn(100) + 50
				p.char = matrixChars[rand.Intn(len(matrixChars))]
				p.color = rand.Intn(255)
			}
		}
	}
	
	return m
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	
	if m.width == 0 || m.height == 0 {
		return "Initializing Terminal Art Studio...\nPress 1-5 to switch modes | q to quit"
	}
	
	var output strings.Builder
	
	switch m.mode {
	case modeMatrix:
		output.WriteString(m.renderMatrix())
	case modeFire:
		output.WriteString(m.renderFire())
	case modeWaves:
		output.WriteString(m.renderWaves())
	case modeParticles:
		output.WriteString(m.renderParticles())
	case modeDrawing:
		output.WriteString(m.renderCanvas())
	}
	
	// Add controls at the bottom
	controls := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("0")).
		Render("1:Matrix  2:Fire  3:Waves  4:Particles  5:Drawing  q:Quit")
	
	return output.String() + "\n" + controls
}

func (m model) renderMatrix() string {
	var output strings.Builder
	
	for y := 0; y < m.height-2; y++ {
		for x := 0; x < m.width; x++ {
			if x < len(m.matrixChars) && y < len(m.matrixChars[x]) {
				intensity := float64(m.height-m.matrixColors[x][y]) / float64(m.height)
				if intensity < 0.1 {
					intensity = 0.1
				}
				
				var color string
				if intensity > 0.8 {
					color = "15" // bright white
				} else if intensity > 0.6 {
					color = "10" // bright green
				} else if intensity > 0.4 {
					color = "2"  // green
				} else {
					color = "22" // dark green
				}
				
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				output.WriteString(style.Render(string(m.matrixChars[x][y])))
			} else {
				output.WriteString(" ")
			}
		}
		if y < m.height-3 {
			output.WriteString("\n")
		}
	}
	
	return output.String()
}

func (m model) renderFire() string {
	var output strings.Builder
	fireChars := []rune{' ', '.', ':', '^', '*', 'x', 's', 'S', '#', '$'}
	
	for y := 0; y < m.height-2; y++ {
		for x := 0; x < m.width; x++ {
			if x < len(m.fireBuffer) && y < len(m.fireBuffer[x]) {
				intensity := m.fireBuffer[x][y]
				charIndex := (intensity * len(fireChars)) / 256
				if charIndex >= len(fireChars) {
					charIndex = len(fireChars) - 1
				}
				
				var color string
				if intensity > 200 {
					color = "15" // white
				} else if intensity > 150 {
					color = "11" // bright yellow
				} else if intensity > 100 {
					color = "3"  // yellow
				} else if intensity > 50 {
					color = "1"  // red
				} else if intensity > 20 {
					color = "9"  // bright red
				} else {
					color = "0"  // black
				}
				
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				output.WriteString(style.Render(string(fireChars[charIndex])))
			} else {
				output.WriteString(" ")
			}
		}
		if y < m.height-3 {
			output.WriteString("\n")
		}
	}
	
	return output.String()
}

func (m model) renderWaves() string {
	var output strings.Builder
	waveChars := []rune{'~', '≈', '∼', '⋉', '⋊', '◊', '♦'}
	
	for y := 0; y < m.height-2; y++ {
		for x := 0; x < m.width; x++ {
			// Create wave effect
			wave1 := math.Sin(float64(x)*0.1+m.wavePhase) * 0.5
			wave2 := math.Sin(float64(y)*0.2+m.wavePhase*0.7) * 0.3
			wave3 := math.Sin((float64(x+y))*0.05+m.wavePhase*1.3) * 0.2
			
			amplitude := wave1 + wave2 + wave3
			intensity := (amplitude + 1.0) / 2.0 // normalize to 0-1
			
			charIndex := int(intensity * float64(len(waveChars)))
			if charIndex >= len(waveChars) {
				charIndex = len(waveChars) - 1
			}
			
			// Color based on wave intensity
			colorVal := int(intensity * 255)
			var color string
			if colorVal > 200 {
				color = "14" // bright cyan
			} else if colorVal > 150 {
				color = "6"  // cyan
			} else if colorVal > 100 {
				color = "4"  // blue
			} else if colorVal > 50 {
				color = "12" // bright blue
			} else {
				color = "8"  // dark gray
			}
			
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			output.WriteString(style.Render(string(waveChars[charIndex])))
		}
		if y < m.height-3 {
			output.WriteString("\n")
		}
	}
	
	return output.String()
}

func (m model) renderParticles() string {
	// Create empty canvas
	canvas := make([][]rune, m.width)
	colors := make([][]string, m.width)
	for i := range canvas {
		canvas[i] = make([]rune, m.height-2)
		colors[i] = make([]string, m.height-2)
		for j := range canvas[i] {
			canvas[i][j] = ' '
			colors[i][j] = "0"
		}
	}
	
	// Draw particles
	for _, p := range m.particles {
		x, y := int(p.x), int(p.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height-2 {
			canvas[x][y] = p.char
			
			// Color based on life
			lifeRatio := float64(p.life) / 150.0
			if lifeRatio > 0.8 {
				colors[x][y] = "15" // white
			} else if lifeRatio > 0.6 {
				colors[x][y] = "13" // bright magenta
			} else if lifeRatio > 0.4 {
				colors[x][y] = "5"  // magenta
			} else if lifeRatio > 0.2 {
				colors[x][y] = "9"  // bright red
			} else {
				colors[x][y] = "1"  // red
			}
		}
	}
	
	// Render canvas
	var output strings.Builder
	for y := 0; y < m.height-2; y++ {
		for x := 0; x < m.width; x++ {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[x][y]))
			output.WriteString(style.Render(string(canvas[x][y])))
		}
		if y < m.height-3 {
			output.WriteString("\n")
		}
	}
	
	return output.String()
}

func (m model) renderCanvas() string {
	var output strings.Builder
	centerX, centerY := m.width/2, (m.height-2)/2
	
	for y := 0; y < m.height-2; y++ {
		for x := 0; x < m.width; x++ {
			// Create a mandala-like pattern
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)
			angle := math.Atan2(dy, dx)
			
			pattern := math.Sin(dist*0.3+float64(m.frame)*0.05) * math.Sin(angle*8)
			if pattern > 0.3 {
				char := '●'
				color := "13" // bright magenta
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				output.WriteString(style.Render(string(char)))
			} else if pattern > 0.1 {
				char := '◆'
				color := "5" // magenta
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				output.WriteString(style.Render(string(char)))
			} else if pattern > -0.1 {
				char := '◇'
				color := "8" // dark gray
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				output.WriteString(style.Render(string(char)))
			} else {
				output.WriteString(" ")
			}
		}
		if y < m.height-3 {
			output.WriteString("\n")
		}
	}
	
	return output.String()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
