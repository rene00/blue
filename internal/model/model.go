package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type Model interface {
	tea.Model
	// Input returns the input that has been received.
	Input() string

	// Reset resets the model.
	Reset()
}

type errMsg error

func NewTextAreaModel() Model {
	t := textarea.New()
	t.Placeholder = "press [tab] to submit prompt"
	t.SetWidth(72)
	t.SetHeight(3)
	t.ShowLineNumbers = false
	t.Focus()
	return &TextAreaModel{
		textarea: t,
	}
}

type TextAreaModel struct {
	textarea textarea.Model
	err      error
	input    string
}

func (m *TextAreaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *TextAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			m.input = m.textarea.Value()
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *TextAreaModel) View() string {
	return fmt.Sprintf(m.textarea.View()) + "\n\n"
}

func (m *TextAreaModel) Input() string {
	return m.input
}

func (m *TextAreaModel) Reset() {
	m.textarea.SetValue("")
}
