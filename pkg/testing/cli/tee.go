/*
Copyright (C) 2017 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

type Tee struct {
	waitGroup    *sync.WaitGroup
	OrigStdout   *os.File
	OrigStderr   *os.File
	Stdout       *os.File
	Stderr       *os.File
	StdoutBuffer *bytes.Buffer
	StderrBuffer *bytes.Buffer
	silent       bool
}

// NewTee splits os.Stdout and os.Stderr using a pipe. While the tee is not closed data
// written to stdout or stderr will be copied into a buffer as well well as printed to the
// original file handle, unless silent is true, in which case the original output streams are silenced.
func NewTee(silent bool) (*Tee, error) {
	origStdout := os.Stdout
	origStderr := os.Stderr

	stdoutRead, stdoutWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	os.Stdout = stdoutWrite
	os.Stderr = stderrWrite

	waitGroup := &sync.WaitGroup{}

	// function to be called as a goroutine that will
	// read from the read-end of our pipe and copy
	// to the buffer as well as writing to the write-end of the pipe
	f := func(reader, writer *os.File, buffer *bytes.Buffer) {
		buf := make([]byte, 4096)
		for true {
			read, err := reader.Read(buf)

			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from buffer: %s", err)
				break
			}

			if read == 0 {
				break
			}

			// write to the original intended output
			if !silent {
				_, err = writer.Write(buf[:read])
				if err != nil {
					break
				}
			}

			_, err = buffer.Write(buf[:read])
			if err != nil {
				break
			}
		}
		waitGroup.Done()
	}
	waitGroup.Add(2)

	var stdoutBuffer bytes.Buffer
	go f(stdoutRead, origStdout, &stdoutBuffer)

	var stderrBuffer bytes.Buffer
	go f(stderrRead, origStderr, &stderrBuffer)
	return &Tee{waitGroup, origStdout, origStderr, stdoutWrite, stderrWrite, &stdoutBuffer, &stderrBuffer, silent}, nil
}

func (t *Tee) Close() {
	t.Stdout.Close()
	t.Stderr.Close()

	t.waitGroup.Wait()

	os.Stdout = t.OrigStdout
	os.Stderr = t.OrigStderr
}
