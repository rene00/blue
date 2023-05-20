package model

import (
	"bytes"
	"regexp"
	"strings"
)

type UserInput struct {
	input    string
	commands []Command
}

func NewUserInput(s string) UserInput {
	u := UserInput{input: s}

	re := regexp.MustCompile(`\bc:([a-z]{1,})$`)
	lines := strings.Split(u.Input(), "\n")
	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) == 2 {
			commandName := match[1]
			switch strings.ToLower(commandName) {
			case "echo":
				u.commands = append(u.commands, NewEchoCommand())
			case "editor":
				u.commands = append(u.commands, NewEditorCommand())
			case "printmessages":
				u.commands = append(u.commands, NewPrintMessagesCommand())
			default:
			}
		}
	}

	return u
}

func (u *UserInput) Update(s string) {
	u.input = s
}

func (u *UserInput) Input() string {
	re := regexp.MustCompile("\n+$")
	s := re.ReplaceAllString(u.input, "")
	return s
}

func (u *UserInput) FilteredInput() string {
	buf := bytes.NewBuffer([]byte{})
	re := regexp.MustCompile(`\bc:([a-z]{1,})$`)
	lines := strings.Split(u.Input(), "\n")
	for _, line := range lines {
		buf.WriteString(re.ReplaceAllString(line, "") + "\n")
	}

	// strip new lines again
	re = regexp.MustCompile("\n+$")
	return re.ReplaceAllString(buf.String(), "")
}

// Commands takes the raw input and appends the commands to a slice.
func (u *UserInput) Commands() []Command {
	return u.commands
}
