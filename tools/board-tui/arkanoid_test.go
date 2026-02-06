package main

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestArkanoidInitAndBounds(t *testing.T) {
	m := newArkanoidModelWithSeed(42)

	if m.fieldWidth < arkanoidMinFieldWidth || m.fieldWidth > arkanoidMaxFieldWidth {
		t.Fatalf("fieldWidth out of bounds: %d", m.fieldWidth)
	}
	if m.fieldHeight < arkanoidMinFieldHeight || m.fieldHeight > arkanoidMaxFieldHeight {
		t.Fatalf("fieldHeight out of bounds: %d", m.fieldHeight)
	}
	if len(m.bricks) == 0 {
		t.Fatal("expected at least one brick")
	}
	if m.paddleX < 0 || m.paddleX > m.fieldWidth-m.paddleWidth {
		t.Fatalf("paddleX out of bounds: %d", m.paddleX)
	}
	if m.ballX < 0 || m.ballX >= m.fieldWidth {
		t.Fatalf("ballX out of bounds: %d", m.ballX)
	}
	if m.ballY < 0 || m.ballY >= m.fieldHeight {
		t.Fatalf("ballY out of bounds: %d", m.ballY)
	}
}

func TestArkanoidBricksDeterministicForSameSeed(t *testing.T) {
	left := newArkanoidModelWithSeed(123)
	right := newArkanoidModelWithSeed(123)

	if !reflect.DeepEqual(left.bricks, right.bricks) {
		t.Fatal("expected identical brick layout for same seed")
	}
}

func TestArkanoidPaddleClamp(t *testing.T) {
	m := newArkanoidModelWithSeed(99)

	m.movePaddle(-1000)
	if m.paddleX != 0 {
		t.Fatalf("expected left clamp at 0, got %d", m.paddleX)
	}

	m.movePaddle(1000)
	want := m.fieldWidth - m.paddleWidth
	if m.paddleX != want {
		t.Fatalf("expected right clamp at %d, got %d", want, m.paddleX)
	}
}

func TestArkanoidKeyMovementStep(t *testing.T) {
	m := newArkanoidModelWithSeed(1234)
	m.paddleX = m.fieldWidth / 2
	start := m.paddleX

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.paddleX != start+arkanoidPaddleStep {
		t.Fatalf("expected right movement by %d, got %d", arkanoidPaddleStep, m.paddleX-start)
	}

	start = m.paddleX
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.paddleX != start-arkanoidPaddleStep {
		t.Fatalf("expected left movement by %d, got %d", arkanoidPaddleStep, start-m.paddleX)
	}
}

func TestArkanoidTrackpadHorizontalSwipeStep(t *testing.T) {
	m := newArkanoidModelWithSeed(4321)
	m.paddleX = m.fieldWidth / 2
	start := m.paddleX

	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelRight})
	if m.paddleX != start+arkanoidPaddleStep {
		t.Fatalf("expected wheel-right movement by %d, got %d", arkanoidPaddleStep, m.paddleX-start)
	}

	start = m.paddleX
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelLeft})
	if m.paddleX != start-arkanoidPaddleStep {
		t.Fatalf("expected wheel-left movement by %d, got %d", arkanoidPaddleStep, start-m.paddleX)
	}
}

func TestArkanoidBrickCollisionScores(t *testing.T) {
	m := newArkanoidModelWithSeed(1)
	m.bricks = []arkanoidBrick{
		{x: 10, y: 5, width: 3, alive: true},
	}
	m.score = 0
	m.won = false
	m.gameOver = false

	m.ballX = 9
	m.ballY = 4
	m.ballDX = 1
	m.ballDY = 1

	m.step()

	if m.score != 10 {
		t.Fatalf("expected score 10, got %d", m.score)
	}
	if m.bricks[0].alive {
		t.Fatal("expected brick to be destroyed")
	}
	if !m.won {
		t.Fatal("expected won=true after clearing final brick")
	}
}

func TestArkanoidLifeLossGameOver(t *testing.T) {
	m := newArkanoidModelWithSeed(7)
	m.bricks = nil
	m.lives = 1
	m.gameOver = false
	m.won = false

	paddleY := m.fieldHeight - 1
	m.paddleX = 0
	m.paddleWidth = 3
	m.ballX = m.fieldWidth - 1
	m.ballY = paddleY
	m.ballDX = 0
	m.ballDY = 1

	m.step()

	if !m.gameOver {
		t.Fatal("expected gameOver=true after losing last life")
	}
	if m.lives != 0 {
		t.Fatalf("expected lives 0, got %d", m.lives)
	}
}

func TestArkanoidZeroValueSetSizeNoPanic(t *testing.T) {
	var m ArkanoidModel

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("SetSize should not panic for zero-value model: %v", recovered)
		}
	}()

	m.SetSize(100, 40)

	if m.rng == nil {
		t.Fatal("expected rng to be initialized")
	}
	if len(m.bricks) == 0 {
		t.Fatal("expected randomized bricks after SetSize")
	}
}
