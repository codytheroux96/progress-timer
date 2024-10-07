package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type inputState int

const (
	inputtingTime inputState = iota
	running
)

type model struct {
	textInput     textinput.Model
	state         inputState
	duration      time.Duration
	timeRemaining time.Duration
	progress      progress.Model
	done          bool
	err           string
}

type tickMsg time.Time

var (
	statusMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)
	completedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter minutes..."
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 20

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
		progress.WithSolidFill("green"),
	)
	
	return model{
		textInput: ti,
		state:     inputtingTime,
		progress:  p,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tickEverySecond(),
		tea.EnterAltScreen,
	)
}

func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.state == inputtingTime {
				input := strings.TrimSpace(m.textInput.Value())
				minutes, err := strconv.Atoi(input)
				if err != nil || minutes <= 0 {
					m.err = "Please enter a valid positive number"
					return m, nil
				}
				m.duration = time.Duration(minutes) * time.Minute
				m.timeRemaining = m.duration
				m.state = running
				m.err = ""
				return m, nil
			}
		}

	case tickMsg:
		if m.state == running && m.timeRemaining > 0 {
			m.timeRemaining -= time.Second
			if m.timeRemaining <= 0 {
				m.timeRemaining = 0
				m.done = true
			}
		}
		return m, tickEverySecond()
	}

	if m.state == inputtingTime {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func (m model) View() string {
	var s strings.Builder

	if m.state == inputtingTime {
		s.WriteString("\nEnter timer duration in minutes:\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		if m.err != "" {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(m.err + "\n\n"))
		}
		s.WriteString("Press Enter to start, Esc to quit\n")
	} else {
		timeStr := formatDuration(m.timeRemaining)
		s.WriteString(fmt.Sprintf("\nTime remaining: %s\n\n", statusMessageStyle.Render(timeStr)))

		elapsed := m.duration - m.timeRemaining
		percentComplete := float64(elapsed) / float64(m.duration)

		m.progress.SetPercent(percentComplete)
		
		progressBar := m.progress.View()
		percentage := fmt.Sprintf("%.1f%%", percentComplete*100)
		
		paddingWidth := 40 - len(percentage)
		padding := strings.Repeat(" ", paddingWidth)
		
		s.WriteString(progressBar)
		s.WriteString(padding)
		s.WriteString(statusMessageStyle.Render(percentage))
		s.WriteString("\n\n")
		
		if m.done {
			s.WriteString(completedStyle.Render("Done!\n\n"))
		}
		
		s.WriteString(fmt.Sprintf("Elapsed: %s / Total: %s\n", 
			formatDuration(elapsed), 
			formatDuration(m.duration)))
		s.WriteString(fmt.Sprintf("Seconds: %.0f / %.0f\n\n", 
			elapsed.Seconds(), 
			m.duration.Seconds()))
		
		s.WriteString("Press Esc to quit\n")
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(s.String())
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}