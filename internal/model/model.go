package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type CommandMsg struct {
	Command Command
}

func NewCommandMsg(c Command) CommandMsg {
	return CommandMsg{Command: c}
}

type errMsg error

type Model struct {
	textarea    textarea.Model
	UserInput   UserInput
	ready       bool
	commandMsgs []CommandMsg
	err         error
}

var initialTextAreaHeight = 3

func NewModel(s string) *Model {
	t := textarea.New()
	t.Placeholder = "press [tab] to submit prompt"
	t.SetWidth(72)
	t.SetHeight(initialTextAreaHeight)
	t.ShowLineNumbers = false
	t.Focus()
	if s != "" {
		t.SetValue(s)
	}

	return &Model{
		textarea: t,
	}
}

func (m *Model) Ready() bool {
	return m.ready
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.textarea.SetHeight(strings.Count(m.textarea.Value(), "\n") + initialTextAreaHeight)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			m.UserInput = NewUserInput(m.textarea.Value())
			m.ready = true

			for _, command := range m.UserInput.Commands() {
				m.commandMsgs = append(m.commandMsgs, NewCommandMsg(command))
			}

			if len(m.commandMsgs) >= 1 {
				m.ready = false
				return m.Update(m.commandMsgs[0])
			}

			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	case CommandMsg:
		command := msg.Command
		if err := command.Run(m); err != nil {
			fmt.Printf("error: %s", err)
			return m, nil
		}

		m.commandMsgs = m.commandMsgs[1:]

		if len(m.commandMsgs) >= 1 {
			return m.Update(m.commandMsgs[0])
		}

		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return fmt.Sprintf(m.textarea.View()) + "\n\n"
}

func (m *Model) Reset() {
	m.textarea.SetValue("")
}
