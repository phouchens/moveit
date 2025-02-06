package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Exercise struct {
	Name        string
	Description string
	Duration    string
}

var exercises = []Exercise{
	{
		Name:        "Push-ups",
		Description: "Place your hands shoulder-width apart, keep your body straight, lower yourself until your chest nearly touches the ground, then push back up.",
		Duration:    "Do 3 sets of 15 repetitions",
	},
	{
		Name:        "Bodyweight Squats",
		Description: "Squat down as deep as possible, Alternatively do wide leg squats.",
		Duration:    "Do 3 sets of 15 repetitions",
	},
	{
		Name:        "Plank",
		Description: "Hold a push-up position with your forearms on the ground. Alternatively- do side planks",
		Duration:    "Hold for 30 seconds, rest, repeat 3 times",
	},
	{
		Name:        "Curls",
		Description: "Grab some dumbbells, Work them guns. If you don't have dumbbells do close grip pushups",
		Duration:    "20-30 reps, 5 sets",
	},
	{
		Name:        "Lunges",
		Description: "Walking Lunges",
		Duration:    "3 sets of 15 each leg",
	},
	{
		Name:        "Overhead Press",
		Description: "Grab some dumbbells, put em overhead",
		Duration:    "5 sets of 15-20",
	},
	{
		Name:        "Side Delt Raise",
		Description: "Grab some dumbbells, Start with hands at side, raise them to the side to slightly above shoulder level, lower controlled.",
		Duration:    "5 sets of 15-20",
	},
}

type model struct {
	progress progress.Model
	start    time.Time
	duration time.Duration
	phase    string
	exercise *Exercise
	quitting bool
	percent  float64
}

func initialModel(duration time.Duration, phase string, exercise *Exercise) model {
	return model{
		progress: progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C")),
		start:    time.Now(),
		duration: duration,
		phase:    phase,
		exercise: exercise,
		percent:  0.0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.EnterAltScreen)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/10, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 20
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil

	case tickMsg:
		elapsed := time.Since(m.start)
		if elapsed >= m.duration {
			return m, tea.Quit
		}

		m.percent = float64(elapsed) / float64(m.duration)
		return m, tickCmd()
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Stay Active!"
	}

	remaining := m.duration - time.Since(m.start)
	str := fmt.Sprintf("\n %s - %s remaining\n\n", m.phase, remaining.Round(time.Second))
	str += " " + m.progress.ViewAs(m.percent) + "\n\n"

	if m.exercise != nil {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		str += style.Render(fmt.Sprintf(" Exercise: %s\n", m.exercise.Name))
		str += fmt.Sprintf(" %s\n", m.exercise.Description)
		str += fmt.Sprintf(" %s\n", m.exercise.Duration)
	}
	str += "\n Press q to quit\n"
	return str
}

func sendMacNotification(title, message string) error {
	message = strings.ReplaceAll(message, `"`, `\"`)
	title = strings.ReplaceAll(title, `"`, `\"`)

	script := fmt.Sprintf(`
        tell application "System Events"
            display notification "%s" with title "%s" subtitle "Workout Timer" sound name "Glass"
        end tell
        tell application "NotificationCenter"
            activate
        end tell`, message, title)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("notification error: %v, output: %s", err, output)
	}
	return nil
}

func runTimer(duration time.Duration, phase string, exercise *Exercise) error {
	p := tea.NewProgram(initialModel(duration, phase, exercise))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("timer error : %v", err)
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <work_duration_minutes> <break_duration_minutes>")
		fmt.Println("Example: go run main.go 25 5")
		os.Exit(1)
	}

	workMinutes, _ := strconv.Atoi(os.Args[1])
	breakMinutes, _ := strconv.Atoi(os.Args[2])

	workDuration := time.Duration(workMinutes) * time.Minute
	breakDuration := time.Duration(breakMinutes) * time.Minute

	for {
		sendMacNotification("Work Period Starting", fmt.Sprintf("Focus for the next %d minutes`", workMinutes))
		if err := runTimer(workDuration, "Work Period", nil); err != nil {
			fmt.Printf("error %v\n", err)
			os.Exit(1)
		}

		exercise := exercises[time.Now().Unix()%int64(len(exercises))]
		sendMacNotification("Break Time", fmt.Sprintf("Time for %s", exercise.Name))
		if err := runTimer(breakDuration, "Break Period", &exercise); err != nil {
			fmt.Printf("error %v\n", err)
			os.Exit(1)
		}

		sendMacNotification("Time to Focus", fmt.Sprint("Good Job, Start Focus Time"))
		fmt.Print("Press Enter to start next session or press 'q' to quit: ")
		var input string
		fmt.Scanln(&input)
		if input == "q" {
			break
		}
	}
}
