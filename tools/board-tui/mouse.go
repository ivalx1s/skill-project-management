package main

import tea "github.com/charmbracelet/bubbletea"

type scrollDirection int

const (
	scrollNone scrollDirection = iota
	scrollUp
	scrollDown
	scrollLeft
	scrollRight
)

// getScrollDirection converts terminal wheel buttons to a semantic direction.
func getScrollDirection(button tea.MouseButton) scrollDirection {
	switch button {
	case tea.MouseButtonWheelUp:
		return scrollUp
	case tea.MouseButtonWheelDown:
		return scrollDown
	case tea.MouseButtonWheelLeft:
		return scrollLeft
	case tea.MouseButtonWheelRight:
		return scrollRight
	default:
		return scrollNone
	}
}

// handleMouseVerticalScroll dispatches wheel events to up/down callbacks.
func handleMouseVerticalScroll(button tea.MouseButton, onUp func(), onDown func()) bool {
	switch getScrollDirection(button) {
	case scrollUp:
		if onUp != nil {
			onUp()
		}
		return true
	case scrollDown:
		if onDown != nil {
			onDown()
		}
		return true
	default:
		return false
	}
}

// handleMouseHorizontalScroll dispatches wheel events to left/right callbacks.
func handleMouseHorizontalScroll(button tea.MouseButton, onLeft func(), onRight func()) bool {
	switch getScrollDirection(button) {
	case scrollLeft:
		if onLeft != nil {
			onLeft()
		}
		return true
	case scrollRight:
		if onRight != nil {
			onRight()
		}
		return true
	default:
		return false
	}
}

// consumeVerticalScrollSteps converts wheel events into signed move steps.
// Positive steps mean scrolling down; negative steps mean scrolling up.
func consumeVerticalScrollSteps(remainder *float64, sensitivity float64, direction scrollDirection) (int, bool) {
	switch direction {
	case scrollUp, scrollDown:
	default:
		return 0, false
	}

	if remainder == nil {
		return 0, true
	}

	scale := ClampScrollSensitivity(sensitivity) / CurrentScrollSpeedSensitivity
	if direction == scrollUp {
		*remainder -= scale
	} else {
		*remainder += scale
	}

	steps := 0
	for *remainder >= 1.0 {
		steps++
		*remainder -= 1.0
	}
	for *remainder <= -1.0 {
		steps--
		*remainder += 1.0
	}

	return steps, true
}
