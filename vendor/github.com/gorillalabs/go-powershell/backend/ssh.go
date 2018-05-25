// Copyright (c) 2017 Gorillalabs. All rights reserved.

package backend

import (
	"io"

	"github.com/juju/errors"
	"golang.org/x/crypto/ssh"
)

type SSH struct {
	Session *ssh.Session
}

func (b *SSH) StartProcess(cmd string, args ...string) (Waiter, io.Writer, io.Reader, io.Reader, error) {
	stdin, err := b.Session.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.Annotate(err, "Could not get hold of the PowerShell's stdin stream")
	}

	stdout, err := b.Session.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.Annotate(err, "Could not get hold of the PowerShell's stdout stream")
	}

	stderr, err := b.Session.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.Annotate(err, "Could not get hold of the PowerShell's stderr stream")
	}

	err = b.Session.Start(cmd) // TODO: quote and add args
	if err != nil {
		return nil, nil, nil, nil, errors.Annotate(err, "Could not spawn PowerShell process")
	}

	return b.Session, stdin, stdout, stderr, nil
}
