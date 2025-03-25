package exec

import "fmt"

type Output struct {
	Command string
	Stdout  string
	Stderr  string
}

func (o Output) AsPrompt() string {
	return fmt.Sprintf(
		"Command: %q\nOutput:\n<stdout>\n%s\n</stdout>\n<stderr>\n%s\n</stderr>",
		o.Command,
		o.Stdout,
		o.Stderr,
	)
}
