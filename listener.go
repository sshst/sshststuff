package client

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	gossh "golang.org/x/crypto/ssh"
)

type listener struct {
	server        *ssh.Server
	command       []string
	activechanged chan int
}

func (l *listener) GetPublicKey() gossh.PublicKey {
	return l.server.HostSigners[0].PublicKey()
}

func (l *listener) SetCertificate(cert *gossh.Certificate) {
	signer, err := gossh.NewCertSigner(cert, l.server.HostSigners[0].(gossh.Signer))
	if err != nil {
		panic(err)
	}

	l.server.HostSigners[0] = signer.(ssh.Signer)
}

func (l *listener) timeout(seconds int) {
	if seconds == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "Will timeout if no connections for %d seconds\n", seconds)
	active := 0

	for {
		select {
		case delta := <-l.activechanged:
			active += delta
		case <-time.After(time.Second * time.Duration(seconds)):
			if active == 0 {
				fmt.Fprintln(os.Stderr, "Exiting gracefully due to inactivity timeout being exceeded")
				os.Exit(0)
			}
		}
	}
}

func shellCommand() []string {
	paths := []string{
		"/bin/bash",
		"/bin/zsh",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return []string{path}
		}
	}

	return []string{"/bin/sh"}
}

func (l *listener) handleSSH(s ssh.Session) {
	fingerprint := fingerprinter(s.PublicKey())

	sconn := s.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	go func() {
		sconn.Wait()
		fmt.Printf("Connection disconnected from %s\n", fingerprint)
		l.activechanged <- -1
	}()

	l.activechanged <- 1
	fmt.Printf("Accepted connection from %s\n", fingerprint)

	command := l.getcmd(s)
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = append(os.Environ(), "SSHST=true")

	ptyReq, winCh, isPty := s.Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			panic(err)
		}

		go handleWinch(winCh, f)
		go io.Copy(f, s)
		io.Copy(s, f) // stdout
	} else {
		out, _ := cmd.StdoutPipe()
		errp, _ := cmd.StderrPipe()
		in, _ := cmd.StdinPipe()
		cmd.Start()

		go io.Copy(s, out)
		go io.Copy(s.Stderr(), errp)
		go func() {
			io.Copy(in, s)
			in.Close()
		}()
		if err := cmd.Wait(); err != nil {
			s.Exit(1)
		} else {
			s.Exit(0)
		}
	}
}

func (l *listener) getcmd(s ssh.Session) []string {
	if len(l.command) > 0 {
		return l.command
	}

	// very loosely based on https://github.com/openssh/openssh-portable/blob/c7c099060f82ffe6a36d8785ecf6052e12fd92f0/session.c#L1680-L1717

	if len(s.Command()) == 0 {
		return shellCommand() // TODO really these need argv0[0] == '-' to be a "login shell"
	}

	return []string{"/bin/sh", "-c", s.RawCommand()}
}

func handleWinch(winCh <-chan ssh.Window, f *os.File) {
	for win := range winCh {
		str := &struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0}
		syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(str)))
	}
}
