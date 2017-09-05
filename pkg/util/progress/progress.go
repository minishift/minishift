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

package progress

import (
	"fmt"
	"time"
)

const DefaultInterval = 2 * time.Second

type ProgressDots struct {
	interval time.Duration
	handler  chan bool
}

// NewProgressDots creates the channel to handle progress dots
func New() *ProgressDots {
	return &ProgressDots{
		interval: DefaultInterval,
		handler:  make(chan bool),
	}
}

// Start starts the dots
func (s *ProgressDots) Start() {
	go func() {
		for {
			select {
			case <-s.handler:
				return
			default:
				fmt.Print(".")
				time.Sleep(s.interval)
			}
		}
	}()
}

// Stop stops the dots
func (s *ProgressDots) Stop() {
	s.handler <- true
}

// SetInterval sets the interval
func (s *ProgressDots) SetInterval(interval time.Duration) {
	s.interval = interval
}
