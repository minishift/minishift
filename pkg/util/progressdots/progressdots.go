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

package progressdots

import (
	"fmt"
	"time"
)

const (
	DefaultInterval = 2 * time.Second
	DefaultEasing   = 10
)

type ProgressDots struct {
	interval   time.Duration
	easing     int
	dotCounter int
	handler    chan bool
}

// New creates the channel to handle progress dots
// It takes easingOptional as optional parameter to control dots interval
func New(easingOptional ...int) *ProgressDots {
	easing := DefaultEasing
	if len(easingOptional) > 0 {
		easing = easingOptional[0]
	}
	return &ProgressDots{
		interval:   DefaultInterval,
		easing:     easing,
		dotCounter: 0,
		handler:    make(chan bool),
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
				s.dotCounter++
				time.Sleep(s.interval)
				if s.easing != 0 && s.dotCounter%s.easing == 0 {
					s.dotCounter = 0
					s.interval = s.interval + 1*time.Second
				}
			}
		}
	}()
}

// Stop stops the dots
func (s *ProgressDots) Stop() {
	defer close(s.handler)
	s.handler <- true
}

// SetInterval sets the interval
func (s *ProgressDots) SetInterval(interval time.Duration) {
	s.interval = interval
}
