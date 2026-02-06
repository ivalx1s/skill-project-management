package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGetScrollDirection(t *testing.T) {
	if getScrollDirection(tea.MouseButtonWheelUp) != scrollUp {
		t.Fatal("expected wheel up to map to scrollUp")
	}
	if getScrollDirection(tea.MouseButtonWheelDown) != scrollDown {
		t.Fatal("expected wheel down to map to scrollDown")
	}
	if getScrollDirection(tea.MouseButtonWheelLeft) != scrollLeft {
		t.Fatal("expected wheel left to map to scrollLeft")
	}
	if getScrollDirection(tea.MouseButtonWheelRight) != scrollRight {
		t.Fatal("expected wheel right to map to scrollRight")
	}
	if getScrollDirection(tea.MouseButtonLeft) != scrollNone {
		t.Fatal("expected non-wheel buttons to map to scrollNone")
	}
}

func TestHandleMouseVerticalScrollDispatch(t *testing.T) {
	var upCount int
	var downCount int

	handled := handleMouseVerticalScroll(tea.MouseButtonWheelUp, func() { upCount++ }, func() { downCount++ })
	if !handled {
		t.Fatal("wheel up should be handled")
	}
	if upCount != 1 || downCount != 0 {
		t.Fatalf("unexpected callback counts after up: up=%d down=%d", upCount, downCount)
	}

	handled = handleMouseVerticalScroll(tea.MouseButtonWheelDown, func() { upCount++ }, func() { downCount++ })
	if !handled {
		t.Fatal("wheel down should be handled")
	}
	if upCount != 1 || downCount != 1 {
		t.Fatalf("unexpected callback counts after down: up=%d down=%d", upCount, downCount)
	}

	handled = handleMouseVerticalScroll(tea.MouseButtonLeft, func() { upCount++ }, func() { downCount++ })
	if handled {
		t.Fatal("left click should not be handled as scroll")
	}
	if upCount != 1 || downCount != 1 {
		t.Fatalf("non-wheel should not trigger callbacks: up=%d down=%d", upCount, downCount)
	}
}

func TestHandleMouseHorizontalScrollDispatch(t *testing.T) {
	var leftCount int
	var rightCount int

	handled := handleMouseHorizontalScroll(tea.MouseButtonWheelLeft, func() { leftCount++ }, func() { rightCount++ })
	if !handled {
		t.Fatal("wheel left should be handled")
	}
	if leftCount != 1 || rightCount != 0 {
		t.Fatalf("unexpected callback counts after left: left=%d right=%d", leftCount, rightCount)
	}

	handled = handleMouseHorizontalScroll(tea.MouseButtonWheelRight, func() { leftCount++ }, func() { rightCount++ })
	if !handled {
		t.Fatal("wheel right should be handled")
	}
	if leftCount != 1 || rightCount != 1 {
		t.Fatalf("unexpected callback counts after right: left=%d right=%d", leftCount, rightCount)
	}

	handled = handleMouseHorizontalScroll(tea.MouseButtonLeft, func() { leftCount++ }, func() { rightCount++ })
	if handled {
		t.Fatal("left click should not be handled as horizontal scroll")
	}
	if leftCount != 1 || rightCount != 1 {
		t.Fatalf("non-wheel should not trigger callbacks: left=%d right=%d", leftCount, rightCount)
	}
}

func TestConsumeVerticalScrollSteps_BaselineMapsToOneStep(t *testing.T) {
	var remainder float64

	steps, handled := consumeVerticalScrollSteps(&remainder, CurrentScrollSpeedSensitivity, scrollDown)
	if !handled {
		t.Fatal("expected scrollDown to be handled")
	}
	if steps != 1 {
		t.Fatalf("expected one step at baseline sensitivity, got %d", steps)
	}
	if remainder != 0 {
		t.Fatalf("expected zero remainder at baseline, got %f", remainder)
	}
}

func TestConsumeVerticalScrollSteps_DefaultSensitivityAccumulates(t *testing.T) {
	var remainder float64

	steps, handled := consumeVerticalScrollSteps(&remainder, DefaultScrollSensitivity, scrollDown)
	if !handled {
		t.Fatal("expected scrollDown to be handled")
	}
	if steps != 0 {
		t.Fatalf("expected zero immediate steps at default sensitivity, got %d", steps)
	}

	steps, handled = consumeVerticalScrollSteps(&remainder, DefaultScrollSensitivity, scrollDown)
	if !handled {
		t.Fatal("expected second scrollDown to be handled")
	}
	if steps <= 0 {
		t.Fatalf("expected accumulated steps to become positive, got %d", steps)
	}
}
