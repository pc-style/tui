# Terminal Art Studio 🎨

An unconventional, stunning TUI application built with [Bubble Tea][bubbletea] that transforms your terminal into a digital art canvas.

## ✨ Features

**Five Breathtaking Visual Modes:**

1. **Matrix Rain** - Cascading Japanese characters with authentic green color gradients
2. **Fire Effect** - Realistic fire simulation with smooth color transitions from red to yellow to white
3. **Wave Patterns** - Hypnotic wave animations using Unicode symbols with cyan/blue gradients
4. **Particle System** - Physics-based floating particles that bounce and fade with vibrant colors
5. **Mandala Drawing** - Animated geometric patterns creating mesmerizing mathematical art

## 🎮 Controls

- **1-5**: Switch between visual modes
- **q/ESC/Ctrl+C**: Quit the application

## 🚀 Getting Started

```bash
go build
./bubbletea-app-template
```

## 🎯 What Makes It Special

- **Unconventional Design**: Unlike practical TUI apps, this is pure digital art
- **Real-time Animation**: Smooth 50ms frame updates for fluid motion
- **Rich Colors**: Leverages full terminal color palette for stunning gradients
- **Unicode Art**: Uses Japanese characters and mathematical symbols for unique aesthetics
- **Responsive**: Adapts to any terminal size dynamically
- **Physics Simulation**: Real particle physics with bouncing and life cycles

## 📦 Dependencies

- [Bubble Tea][bubbletea] - The TUI framework
- [Lip Gloss][lipgloss] - Terminal styling and colors

## 🎨 Technical Details

Each visual mode implements different algorithms:
- **Matrix**: Character streaming with aging/fading effects
- **Fire**: Cellular automata-based fire propagation 
- **Waves**: Multiple sine wave interference patterns
- **Particles**: 2D physics simulation with collision detection
- **Mandala**: Real-time mathematical pattern generation

[bubbletea]: https://github.com/charmbracelet/bubbletea
[lipgloss]: https://github.com/charmbracelet/lipgloss
