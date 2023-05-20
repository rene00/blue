package model

import "testing"

func TestUserInputInput(t *testing.T) {

	t.Run("Input()", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{
				"foo\n\n",
				"foo",
			},
			{
				"foo\nfoo",
				"foo\nfoo",
			},
			{
				"foo\nfoo\n",
				"foo\nfoo",
			},
			{
				"foo",
				"foo",
			},
		}

		for _, test := range tests {
			i := NewUserInput(test.input)
			if got := i.Input(); got != test.want {
				t.Errorf("failed; got=%s want=%s", got, test.want)
			}
		}
	})
}

func TestUserInputCommands(t *testing.T) {
	t.Run("Commands()", func(t *testing.T) {
		tests := []struct {
			input string
			want  []Command
		}{
			{
				"foo\n\n",
				[]Command{},
			},
			{
				"foo1\nc:echo\n",
				[]Command{NewEchoCommand()},
			},
			{
				"foo2\nc:echo\nc:echo",
				[]Command{NewEchoCommand(), NewEchoCommand()},
			},
			{
				// c:echo needs to exist on it's own line.
				"foo2c:echo\n",
				[]Command{},
			},
		}

		for _, test := range tests {
			i := NewUserInput(test.input)
			got := i.Commands()
			r := compareCommandSlices(got, test.want)
			if !r {
				t.Errorf("failed; got=%s want=%s", got, test.want)
			}
		}
	})
}

func compareCommandSlices(c1, c2 []Command) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := 0; i < len(c1); i++ {
		if c1[i] != c2[i] {
			return false
		}
	}
	return true
}
