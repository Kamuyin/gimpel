package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	gimpelsdk "github.com/NoHaxxJustLags/gimpel/sdk/go"
)

type SSHHoneypot struct {
	ctx      *gimpelsdk.ModuleContext
	emitter  *gimpelsdk.LocalEventEmitter
	config   *ssh.ServerConfig
	hostKey  ssh.Signer
	sessions sync.Map
}

func NewSSHHoneypot() *SSHHoneypot {
	return &SSHHoneypot{}
}

func (h *SSHHoneypot) Name() string {
	return "ssh-honeypot"
}

func (h *SSHHoneypot) Init(ctx *gimpelsdk.ModuleContext) error {
	h.ctx = ctx
	h.emitter = gimpelsdk.NewLocalEventEmitter(ctx.ModuleID, 1000)
	ctx.Emitter = h.emitter

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating host key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return fmt.Errorf("creating signer: %w", err)
	}
	h.hostKey = signer

	h.config = &ssh.ServerConfig{
		PasswordCallback:  h.handlePasswordAuth,
		PublicKeyCallback: h.handlePublicKeyAuth,
		ServerVersion:     "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.1",
	}
	h.config.AddHostKey(h.hostKey)

	go h.processEvents()

	log.Printf("SSH honeypot initialized")
	return nil
}

func (h *SSHHoneypot) HandleConnection(ctx context.Context, conn net.Conn, info *gimpelsdk.ConnectionInfo) error {
	sessionID := uuid.New().String()

	h.emitter.EmitConnectionOpen(
		sessionID,
		info.SourceIP,
		info.DestIP,
		"ssh",
		info.SourcePort,
		info.DestPort,
	)

	defer func() {
		h.emitter.EmitConnectionClose(sessionID)
		h.sessions.Delete(sessionID)
	}()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, h.config)
	if err != nil {
		log.Printf("SSH handshake failed from %s: %v", info.SourceIP, err)
		return nil
	}
	defer sshConn.Close()

	h.sessions.Store(sessionID, sshConn)

	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChan.Accept()
		if err != nil {
			continue
		}

		go h.handleSession(ctx, sessionID, channel, requests)
	}

	return nil
}

func (h *SSHHoneypot) handlePasswordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	log.Printf("Password auth attempt: user=%s password=%s from=%s",
		conn.User(), string(password), conn.RemoteAddr())

	h.emitter.EmitAuthAttempt("", map[string]string{
		"method":   "password",
		"username": conn.User(),
		"password": string(password),
		"remote":   conn.RemoteAddr().String(),
	})

	return &ssh.Permissions{
		Extensions: map[string]string{
			"user": conn.User(),
		},
	}, nil
}

func (h *SSHHoneypot) handlePublicKeyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Printf("Public key auth attempt: user=%s key_type=%s from=%s",
		conn.User(), key.Type(), conn.RemoteAddr())

	h.emitter.EmitAuthAttempt("", map[string]string{
		"method":      "publickey",
		"username":    conn.User(),
		"key_type":    key.Type(),
		"fingerprint": ssh.FingerprintSHA256(key),
		"remote":      conn.RemoteAddr().String(),
	})

	return &ssh.Permissions{
		Extensions: map[string]string{
			"user": conn.User(),
		},
	}, nil
}

func (h *SSHHoneypot) handleSession(ctx context.Context, sessionID string, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	var shell bool

	for req := range requests {
		switch req.Type {
		case "shell":
			shell = true
			req.Reply(true, nil)
		case "pty-req":
			req.Reply(true, nil)
		case "exec":
			if len(req.Payload) > 4 {
				cmdLen := uint32(req.Payload[0])<<24 | uint32(req.Payload[1])<<16 | uint32(req.Payload[2])<<8 | uint32(req.Payload[3])
				if int(cmdLen) <= len(req.Payload)-4 {
					cmd := req.Payload[4 : 4+cmdLen]
					log.Printf("Exec command: %s", string(cmd))
					h.emitter.EmitCommand(sessionID, cmd)
				}
			}
			req.Reply(true, nil)
			channel.Write([]byte("Command executed\n"))
			channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			return
		case "env":
			req.Reply(true, nil)
		default:
			req.Reply(false, nil)
		}
	}

	if shell {
		h.runFakeShell(ctx, sessionID, channel)
	}
}

func (h *SSHHoneypot) runFakeShell(ctx context.Context, sessionID string, channel ssh.Channel) {
	prompt := "root@honeypot:~# "
	channel.Write([]byte("Welcome to Ubuntu 22.04.2 LTS\n\n"))
	channel.Write([]byte(prompt))

	buf := make([]byte, 1024)
	var cmdBuf []byte

	for {
		n, err := channel.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			return
		}

		for i := 0; i < n; i++ {
			b := buf[i]

			switch b {
			case '\r', '\n':
				channel.Write([]byte("\r\n"))
				if len(cmdBuf) > 0 {
					cmd := string(cmdBuf)
					log.Printf("Shell command: %s", cmd)
					h.emitter.EmitCommand(sessionID, cmdBuf)
					h.executeCommand(channel, cmd)
					cmdBuf = nil
				}
				channel.Write([]byte(prompt))

			case 127, 8:
				if len(cmdBuf) > 0 {
					cmdBuf = cmdBuf[:len(cmdBuf)-1]
					channel.Write([]byte("\b \b"))
				}

			case 3:
				channel.Write([]byte("^C\r\n" + prompt))
				cmdBuf = nil

			case 4:
				channel.Write([]byte("logout\r\n"))
				return

			default:
				if b >= 32 && b < 127 {
					cmdBuf = append(cmdBuf, b)
					channel.Write([]byte{b})
				}
			}
		}
	}
}

func (h *SSHHoneypot) executeCommand(channel ssh.Channel, cmd string) {
	responses := map[string]string{
		"ls":              "Desktop  Documents  Downloads  Music  Pictures  Videos\n",
		"pwd":             "/root\n",
		"whoami":          "root\n",
		"id":              "uid=0(root) gid=0(root) groups=0(root)\n",
		"uname -a":        "Linux honeypot 5.15.0-76-generic #83-Ubuntu SMP x86_64 GNU/Linux\n",
		"cat /etc/passwd": "root:x:0:0:root:/root:/bin/bash\ndaemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n",
		"exit":            "",
	}

	if resp, ok := responses[cmd]; ok {
		if cmd == "exit" {
			return
		}
		channel.Write([]byte(resp))
	} else {
		channel.Write([]byte(fmt.Sprintf("bash: %s: command not found\n", cmd)))
	}
}

func (h *SSHHoneypot) HealthCheck(ctx context.Context) (bool, string) {
	return true, "running"
}

func (h *SSHHoneypot) Shutdown(ctx context.Context) error {
	h.sessions.Range(func(key, value interface{}) bool {
		if conn, ok := value.(*ssh.ServerConn); ok {
			conn.Close()
		}
		return true
	})

	h.emitter.Close()
	log.Printf("SSH honeypot shutdown complete")
	return nil
}

func (h *SSHHoneypot) processEvents() {
	for event := range h.emitter.Events() {
		log.Printf("Event: type=%d session=%s labels=%v",
			event.Type, event.SessionID, event.Labels)
	}
}
