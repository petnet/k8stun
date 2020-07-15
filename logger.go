package k8stun

import (
	"fmt"
	"log"
	"strings"

	au "github.com/logrusorgru/aurora"
)

type colorifier func(arg interface{}) au.Value

var colorifiers = []colorifier{
	au.Green,
	au.BrightYellow,
	au.Blue,
	au.BrightMagenta,
	au.Cyan,
	au.BrightGreen,
	au.Yellow,
	au.BrightBlue,
	au.Magenta,
	au.BrightCyan,
}

// Logger implements formatted logging for tunnels via io.Writer
type Logger struct {
	t     *Tunnel
	Label string
}

func (tl *Logger) colorize(arg interface{}) au.Value {
	f := colorifiers[tl.t.id%len(colorifiers)]
	return f(arg)
}

// Write implements io.Writer
func (tl *Logger) Write(out []byte) (int, error) {
	msg := strings.TrimSpace(string(out))
	line := tl.colorize(fmt.Sprintf("%s %s> %s", tl.Label, tl.t.Name, msg))
	log.Print(line)
	return len(line.String()), nil
}

// Printf provides string formatted logging
func (tl *Logger) Printf(format string, a ...interface{}) {
	tl.Write([]byte(fmt.Sprintf(format, a...)))
}
