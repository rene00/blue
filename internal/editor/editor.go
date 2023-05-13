package editor

import (
	"os"
	"os/exec"
)

type Editor struct {
	Opts
}

type Opts struct {
	// The basename of the editor (i.e. vim)
	basename string

	// The initial content of the file when opening it.
	initialContent string
}

type OptFunc func(*Opts)

func defaultOpts() Opts {
	return Opts{
		basename:       "",
		initialContent: "",
	}
}

func WithBasename(b string) OptFunc {
	return func(opts *Opts) {
		opts.basename = b
	}
}

func WithInitialContent(i string) OptFunc {
	return func(opts *Opts) {
		opts.initialContent = i
	}
}

func NewEditor(opts ...OptFunc) Editor {
	o := defaultOpts()

	for _, fn := range opts {
		fn(&o)
	}

	if basename, ok := os.LookupEnv("EDITOR"); ok && o.basename == "" {
		o.basename = basename
	}

	return Editor{Opts: o}
}

func (e Editor) createTempFile() (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return tmpFile, err
	}

	_, err = tmpFile.WriteString(e.initialContent)
	if err != nil {
		return tmpFile, err
	}
	defer tmpFile.Close()

	return tmpFile, nil
}

func (e Editor) Edit() (string, error) {
	content := ""

	tmpFile, err := e.createTempFile()
	if err != nil {
		return content, err
	}

	defer os.Remove(tmpFile.Name())

	c := exec.Command(e.basename, tmpFile.Name())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		return content, err
	}

	f, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return content, err
	}

	content = string(f)
	return content, nil
}
