package model

import (
	"blue/internal/editor"
	"fmt"
)

type Command interface {
	Run(m *Model) error
}

func NewEchoCommand() Command {
	return EchoCommand{}
}

type EchoCommand struct{}

func (c EchoCommand) Run(m *Model) error {
	fmt.Printf("%s\n\n", m.UserInput.FilteredInput())
	return nil
}

type EditorCommand struct{}

func NewEditorCommand() Command {
	return EditorCommand{}
}

func (c EditorCommand) Run(m *Model) error {
	e := editor.NewEditor(editor.WithInitialContent(m.UserInput.FilteredInput()))
	content, err := e.Edit()
	if err != nil {
		return err
	}
	m.UserInput.Update(content)
	return nil
}

type PrintMessagesCommand struct{}

func NewPrintMessagesCommand() Command {
	return PrintMessagesCommand{}
}

func (c PrintMessagesCommand) Run(m *Model) error {
	fmt.Printf("%#+v\n\n", m.Messages())
	return nil
}
