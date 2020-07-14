package k8stun

import (
	"fmt"
	"log"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

var colors = []string{Green, Yellow, Blue, Purple, Cyan, White, Gray}

// Logger implements formatted logging for tunnels via io.Writer
type Logger struct {
	t     *Tunnel
	Label string
}

func (tl *Logger) color() string {
	return colors[tl.t.id%len(colors)]
}

// Write implements io.Writer
func (tl *Logger) Write(out []byte) (int, error) {
	msg := strings.TrimSpace(string(out))
	log.Printf("%s %s %s> %s %s",
		tl.color(),
		tl.t.Name, tl.Label, msg,
		Reset,
	)
	return len(out), nil
}

// Printf provides string formatted logging
func (tl *Logger) Printf(format string, a ...interface{}) {
	tl.Write([]byte(fmt.Sprintf(format, a...)))
}
