package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"time"
)

type SshClient struct {
	sshInput  io.WriteCloser
	sshOutput *bytes.Buffer
	session   *ssh.Session
	client    *ssh.Client
}

func Connect(ip, user, password string) (*SshClient, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	client, err := ssh.Dial("tcp", ip+":ssh", config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}

	sshOutput := new(bytes.Buffer)
	session.Stdout, session.Stderr = sshOutput, sshOutput
	sshInput, _ := session.StdinPipe()

	modes := ssh.TerminalModes{ssh.ECHO: 0, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := session.RequestPty("Xterm", 80, 40, modes); err != nil {
		return nil, fmt.Errorf("failed to request pty: %s", err)
	}

	if err := session.Shell(); err != nil {
		return nil, fmt.Errorf("failed to start shell: %s", err)
	}
	c := &SshClient{sshInput: sshInput, sshOutput: sshOutput, session: session, client: client}
	c.RecvUntil("~]", 3)

	return c, nil
}

func (c *SshClient) Disconnect() {
	c.client.Close()
	c.session.Close()
}

func (c *SshClient) Send(cmd, expected string, timeout time.Duration) (res []byte, err error) {
	if err = c.SendCommand(cmd); err != nil {
		return nil, err
	}

	return c.RecvUntil(expected, timeout)
}

func (c *SshClient) SendCommand(cmd string) error {
	_, err := fmt.Fprint(c.sshInput, cmd, "\n")
	return err
}

func (c *SshClient) RecvUntil(expected string, timeout time.Duration) (res []byte, err error) {
	var end = time.Now().Add(timeout * time.Second)
	var exp = []byte(expected)

	ticker := time.NewTicker(10 * time.Millisecond)
	for range ticker.C {
		b := c.sshOutput.Bytes()

		if index := bytes.LastIndex(b, exp); index != -1 {
			n := index + len(exp)
			res = make([]byte, n)
			copy(res, c.sshOutput.Next(n))
			break
		}

		if time.Now().After(end) {
			res = make([]byte, len(b))
			copy(res, b)
			err = errors.New("Cann't recv the expected string")
			break
		}
	}

	ticker.Stop()
	return res, err
}
