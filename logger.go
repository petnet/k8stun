package k8stun

import (
	"fmt"
	"log"
)

// Logger implements formatted logging via io.Writer
type Logger struct {
	t     *Tunnel
	Label string
}

// Write implements io.Writer
func (tl *Logger) Write(out []byte) (int, error) {
	log.Printf("%s %s> %s\n",
		tl.t.Name, tl.Label, string(out),
	)
	return len(out), nil
}

// Printf provides string formatted logging
func (tl *Logger) Printf(format string, a ...interface{}) {
	tl.Write([]byte(fmt.Sprintf(format, a...)))
}
