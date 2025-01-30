package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestInitialModel tests the creation of the initial model
func TestInitialModel(t *testing.T) {
	duration := 5 * time.Minute
	phase := "Work Period"
	exercise := &Exercise{
		Name:        "Test Exercise",
		Description: "Test Description",
		Duration:    "Test Duration",
	}

	model := initialModel(duration, phase, exercise)

	assert.Equal(t, duration, model.duration)
	assert.Equal(t, phase, model.phase)
	assert.Equal(t, exercise, model.exercise)
	assert.False(t, model.quitting)
	assert.Equal(t, 0.0, model.percent)
}

// TestModelUpdate tests various model update scenarios
func TestModelUpdate(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		model    model
		wantQuit bool
	}{
		{
			name:     "quit on q press",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			model:    initialModel(time.Minute, "test", nil),
			wantQuit: true,
		},
		{
			name:     "quit on ctrl+c",
			msg:      tea.KeyMsg{Type: tea.KeyCtrlC},
			model:    initialModel(time.Minute, "test", nil),
			wantQuit: true,
		},
		{
			name:     "continue on other key",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			model:    initialModel(time.Minute, "test", nil),
			wantQuit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := tt.model.Update(tt.msg)
			m := updatedModel.(model)

			assert.Equal(t, tt.wantQuit, m.quitting)
			if tt.wantQuit {
				assert.NotNil(t, cmd)
			}
		})
	}
}

// TestExerciseSelection tests the exercise selection logic
func TestExerciseSelection(t *testing.T) {
	// Test that we get valid exercises
	for i := 0; i < 10; i++ {
		exercise := exercises[time.Now().Unix()%int64(len(exercises))]
		assert.NotEmpty(t, exercise.Name)
		assert.NotEmpty(t, exercise.Description)
		assert.NotEmpty(t, exercise.Duration)
	}
}

// TestProgressCalculation tests the progress calculation
func TestProgressCalculation(t *testing.T) {
	m := initialModel(10*time.Second, "test", nil)
	m.start = time.Now().Add(-5 * time.Second) // simulate 5 seconds passed

	updatedModel, _ := m.Update(tickMsg{})
	newModel := updatedModel.(model)

	// Progress should be around 0.5 (50%)
	assert.InDelta(t, 0.5, newModel.percent, 0.1)
}

// TestView tests the view generation
func TestView(t *testing.T) {
	exercise := &Exercise{
		Name:        "Test Exercise",
		Description: "Test Description",
		Duration:    "Test Duration",
	}

	m := initialModel(time.Minute, "Test Phase", exercise)

	view := m.View()

	// Check that all important elements are present
	assert.Contains(t, view, "Test Phase")
	assert.Contains(t, view, "Test Exercise")
	assert.Contains(t, view, "Test Description")
	assert.Contains(t, view, "Test Duration")
	assert.Contains(t, view, "Press q to quit")
}

// TestNotificationFormatting tests the notification message formatting
// func TestNotificationFormatting(t *testing.T) {
// 	title := `Test "title" with quotes`
// 	message := `Test "message" with quotes`
//
// 	err := sendMacNotification(title, message)
// 	assert.NoError(t, err)
// }
//
// // TestTimerFlow tests the overall timer flow
// func TestTimerFlow(t *testing.T) {
// 	// Short duration for testing
// 	duration := 100 * time.Millisecond
//
// 	err := runTimer(duration, "Test Phase", nil)
// 	assert.NoError(t, err)
// }

// TestWindowResize tests window resize handling
func TestWindowResize(t *testing.T) {
	m := initialModel(time.Minute, "test", nil)

	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}

	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(model)

	assert.Equal(t, 80, newModel.progress.Width) // Should be capped at 80
}
