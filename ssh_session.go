// reverseSSH - a lightweight ssh server with a reverse connection feature
// Copyright (C) 2021  Ferdinor <ferdinor@mailbox.org>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"io"
	"log"
	"os/exec"

	"github.com/gliderlabs/ssh"
)

func makeSSHSessionHandler(shell string) ssh.Handler {
	return func(s ssh.Session) {
		log.Printf("New login from %s@%s", s.User(), s.RemoteAddr().String())
		_, _, isPty := s.Pty()

		switch {
		case isPty:
			log.Println("PTY requested")

			createPty(s, shell)

		case len(s.Command()) > 0:
			log.Printf("No PTY requested, executing command: '%s'", s.RawCommand())

			cmd := exec.Command(s.Command()[0], s.Command()[1:]...)
			// We use StdinPipe to avoid blocking on missing input
			if stdIn, err := cmd.StdinPipe(); err != nil {
				log.Println("Could not initialize StdInPipe", err)
				s.Exit(1)
			} else {
				go io.Copy(stdIn, s)
			}
			cmd.Stdout = s
			cmd.Stderr = s

			if err := cmd.Run(); err != nil {
				log.Println("Command execution failed:", err)
				io.WriteString(s, err.Error())
			}
			s.Exit(cmd.ProcessState.ExitCode())

		default:
			log.Println("No PTY requested, no command supplied")

			select {
			case <-s.Context().Done():
				log.Println("Session closed")
			}
		}
	}
}
