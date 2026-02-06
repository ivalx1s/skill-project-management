package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	arkanoidFrameDuration  = 45 * time.Millisecond
	arkanoidPaddleStep     = 5
	arkanoidMinFieldWidth  = 36
	arkanoidMaxFieldWidth  = 96
	arkanoidMinFieldHeight = 16
	arkanoidMaxFieldHeight = 30
)

type arkanoidFrameMsg struct{}

type arkanoidBrick struct {
	x     int
	y     int
	width int
	alive bool
}

// ArkanoidModel renders and updates a small Arkanoid mini-game screen.
type ArkanoidModel struct {
	width       int
	height      int
	fieldWidth  int
	fieldHeight int

	paddleX     int
	paddleWidth int

	ballX  int
	ballY  int
	ballDX int
	ballDY int

	bricks []arkanoidBrick

	score    int
	lives    int
	gameOver bool
	won      bool

	rng *rand.Rand
}

func (m *ArkanoidModel) ensureRNG() {
	if m.rng == nil {
		m.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

func NewArkanoidModel() ArkanoidModel {
	return newArkanoidModelWithSeed(time.Now().UnixNano())
}

func newArkanoidModelWithSeed(seed int64) ArkanoidModel {
	m := ArkanoidModel{
		paddleWidth: 9,
		rng:         rand.New(rand.NewSource(seed)),
	}
	m.SetSize(80, 24)
	return m
}

func (m *ArkanoidModel) SetSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	m.width = width
	m.height = height

	fieldWidth := clampInt(width-8, arkanoidMinFieldWidth, arkanoidMaxFieldWidth)
	fieldHeight := clampInt(height-10, arkanoidMinFieldHeight, arkanoidMaxFieldHeight)

	if m.fieldWidth == fieldWidth && m.fieldHeight == fieldHeight && len(m.bricks) > 0 {
		m.paddleX = clampInt(m.paddleX, 0, maxInt(0, m.fieldWidth-m.paddleWidth))
		return
	}

	m.fieldWidth = fieldWidth
	m.fieldHeight = fieldHeight
	m.resetGame()
}

func (m ArkanoidModel) Start() tea.Cmd {
	return m.frameCmd()
}

func (m ArkanoidModel) frameCmd() tea.Cmd {
	return tea.Tick(arkanoidFrameDuration, func(time.Time) tea.Msg {
		return arkanoidFrameMsg{}
	})
}

func (m ArkanoidModel) Update(msg tea.Msg) (ArkanoidModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.movePaddle(-arkanoidPaddleStep)
		case "right", "l":
			m.movePaddle(arkanoidPaddleStep)
		case "r":
			m.resetGame()
		}
		return m, nil

	case tea.MouseMsg:
		if handleMouseHorizontalScroll(msg.Button,
			func() { m.movePaddle(-arkanoidPaddleStep) },
			func() { m.movePaddle(arkanoidPaddleStep) },
		) {
			return m, nil
		}
		return m, nil

	case arkanoidFrameMsg:
		if !m.gameOver && !m.won {
			m.step()
		}
		return m, m.frameCmd()
	}

	return m, nil
}

func (m ArkanoidModel) View() string {
	title := titleStyle.Render(" Arkanoid ")
	status := statusBarStyle.Render(fmt.Sprintf(" Score: %d | Lives: %d | Bricks: %d ", m.score, m.lives, m.remainingBricks()))
	field := m.renderField()
	help := helpStyle.Render("  ←/→: move | trackpad ⇠/⇢ swipe | r: restart | esc: back")

	state := ""
	if m.won {
		state = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")).Render("  You cleared all bricks. Press r to reshuffle.")
	} else if m.gameOver {
		state = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Render("  Game over. Press r to restart.")
	}

	if state != "" {
		return appStyle.Render(title + status + "\n\n" + field + "\n" + state + "\n" + help)
	}
	return appStyle.Render(title + status + "\n\n" + field + "\n" + help)
}

func (m *ArkanoidModel) resetGame() {
	m.ensureRNG()

	m.score = 0
	m.lives = 3
	m.gameOver = false
	m.won = false

	if m.paddleWidth <= 0 {
		m.paddleWidth = 9
	}
	if m.paddleWidth >= m.fieldWidth {
		m.paddleWidth = maxInt(4, m.fieldWidth-2)
	}
	m.paddleX = (m.fieldWidth - m.paddleWidth) / 2

	m.generateBricks()
	m.resetBall()
}

func (m *ArkanoidModel) resetBall() {
	m.ensureRNG()

	m.ballX = m.fieldWidth / 2
	m.ballY = m.fieldHeight - 3
	m.ballDY = -1
	if m.rng.Intn(2) == 0 {
		m.ballDX = -1
	} else {
		m.ballDX = 1
	}
}

func (m *ArkanoidModel) movePaddle(delta int) {
	m.paddleX = clampInt(m.paddleX+delta, 0, maxInt(0, m.fieldWidth-m.paddleWidth))
}

func (m *ArkanoidModel) step() {
	nextX := m.ballX + m.ballDX
	nextY := m.ballY + m.ballDY

	if nextX < 0 || nextX >= m.fieldWidth {
		m.ballDX *= -1
		nextX = m.ballX + m.ballDX
	}

	if nextY < 0 {
		m.ballDY = 1
		nextY = m.ballY + m.ballDY
	}

	paddleY := m.fieldHeight - 1
	if m.ballDY > 0 && nextY == paddleY {
		if nextX >= m.paddleX && nextX < m.paddleX+m.paddleWidth {
			m.ballDY = -1
			hitOffset := nextX - m.paddleX
			third := maxInt(1, m.paddleWidth/3)
			switch {
			case hitOffset < third:
				m.ballDX = -1
			case hitOffset >= 2*third:
				m.ballDX = 1
			}
			nextY = paddleY - 1
		}
	}

	if nextY > paddleY {
		m.lives--
		if m.lives <= 0 {
			m.gameOver = true
			return
		}
		m.resetBall()
		return
	}

	if m.hitBrick(nextX, nextY) {
		m.ballDY *= -1
		nextY = m.ballY + m.ballDY
		if m.remainingBricks() == 0 {
			m.won = true
		}
	}

	m.ballX = nextX
	m.ballY = nextY
}

func (m *ArkanoidModel) hitBrick(ballX, ballY int) bool {
	for i := range m.bricks {
		brick := &m.bricks[i]
		if !brick.alive {
			continue
		}
		if ballY == brick.y && ballX >= brick.x && ballX < brick.x+brick.width {
			brick.alive = false
			m.score += 10
			return true
		}
	}
	return false
}

func (m *ArkanoidModel) generateBricks() {
	m.ensureRNG()

	m.bricks = m.bricks[:0]

	rows := 6
	if m.fieldHeight < 20 {
		rows = 4
	}
	cols := maxInt(6, (m.fieldWidth-6)/4)
	totalWidth := cols*4 - 1
	startX := (m.fieldWidth - totalWidth) / 2
	if startX < 1 {
		startX = 1
	}

	aliveCount := 0
	for y := 0; y < rows; y++ {
		for col := 0; col < cols; col++ {
			include := m.rng.Float64() < 0.68
			if !include && (y+col)%7 == 0 {
				include = true
			}
			if !include {
				continue
			}

			m.bricks = append(m.bricks, arkanoidBrick{
				x:     startX + col*4,
				y:     1 + y,
				width: 3,
				alive: true,
			})
			aliveCount++
		}
	}

	if aliveCount == 0 {
		m.bricks = append(m.bricks, arkanoidBrick{
			x:     maxInt(1, m.fieldWidth/2-1),
			y:     1,
			width: 3,
			alive: true,
		})
	}
}

func (m ArkanoidModel) renderField() string {
	grid := make([][]rune, m.fieldHeight)
	for y := 0; y < m.fieldHeight; y++ {
		row := make([]rune, m.fieldWidth)
		for x := range row {
			row[x] = ' '
		}
		grid[y] = row
	}

	for _, brick := range m.bricks {
		if !brick.alive || brick.y < 0 || brick.y >= m.fieldHeight {
			continue
		}
		for i := 0; i < brick.width; i++ {
			x := brick.x + i
			if x >= 0 && x < m.fieldWidth {
				grid[brick.y][x] = '█'
			}
		}
	}

	paddleY := m.fieldHeight - 1
	for i := 0; i < m.paddleWidth; i++ {
		x := m.paddleX + i
		if x >= 0 && x < m.fieldWidth {
			grid[paddleY][x] = '='
		}
	}

	if !m.gameOver && m.ballY >= 0 && m.ballY < m.fieldHeight && m.ballX >= 0 && m.ballX < m.fieldWidth {
		grid[m.ballY][m.ballX] = '●'
	}

	var sb strings.Builder
	sb.WriteString("┌" + strings.Repeat("─", m.fieldWidth) + "┐\n")
	for _, row := range grid {
		sb.WriteRune('│')
		sb.WriteString(string(row))
		sb.WriteRune('│')
		sb.WriteRune('\n')
	}
	sb.WriteString("└" + strings.Repeat("─", m.fieldWidth) + "┘")

	return sb.String()
}

func (m ArkanoidModel) remainingBricks() int {
	alive := 0
	for _, brick := range m.bricks {
		if brick.alive {
			alive++
		}
	}
	return alive
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
