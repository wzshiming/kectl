/*
Copyright 2024 The Kubernetes Authors.

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

package handle

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/wzshiming/getch"
)

// Handle is a struct that represents a handle with pause and Speed properties.
type Handle struct {
	// pause is a boolean that represents whether the handle is paused or not.
	pause atomic.Bool
	// speed is a pointer to a speed object that represents the speed of the handle.
	speed atomic.Pointer[Speed]
}

// NewHandle creates a new handle.
func NewHandle() *Handle {
	h := &Handle{
		pause: atomic.Bool{},
		speed: atomic.Pointer[Speed]{},
	}

	var s = Speed(1)
	h.speed.Store(&s)
	return h
}

// Input is handles the input from the user.
func (h *Handle) Input(ctx context.Context) {
	for ctx.Err() == nil {
		r, _, err := getch.Getch()
		if err != nil {
			slog.Error("Failed to get key", "err", err)
			return
		}
		switch r {
		case getch.Key_u, getch.KeyU:
			s := h.SpeedUp()
			slog.Info("Speed up", "rate", s)
		case getch.Key_d, getch.KeyD:
			s := h.SpeedDown()
			slog.Info("Speed down", "rate", s)
		case getch.KeySpace:
			if !h.IsPause() {
				h.Pause()
				slog.Info("Paused, Press `Enter` key to continue")
			} else {
				slog.Info("Already paused, Press `Enter` key to continue")
			}
		case getch.KeyCtrlJ: // Enter
			if h.IsPause() {
				h.Continue()
				slog.Info("Continue, Press `Space` key to pause")
			} else {
				slog.Info("Already running, Press `Space` key to pause")
			}
		default:
			slog.Warn("Unknown key", "key", r)
		}
	}
}

// Info is logs the instructions for the user.
func (h *Handle) Info(ctx context.Context) {
	slog.Info("Press `Space` key to pause, press `Enter` key to continue")
	slog.Info("Press `U` key to speed up, press `D` key to speed down")
}

// Speed is returns the speed of the handle.
func (h *Handle) Speed() Speed {
	return *h.speed.Load()
}

// SpeedUp is increases the speed of the handle.
func (h *Handle) SpeedUp() Speed {
	s := *h.speed.Load()
	sd := s.Up()
	if sd <= 10 {
		h.speed.Store(&sd)
		return sd
	}
	return s
}

// SpeedDown is decreases the speed of the handle.
func (h *Handle) SpeedDown() Speed {
	s := *h.speed.Load()
	sd := s.Down()
	if sd != 0 {
		h.speed.Store(&sd)
		return sd
	}
	return s
}

// IsPause is returns whether the handle is paused or not.
func (h *Handle) IsPause() bool {
	return h.pause.Load()
}

// Pause is pauses the handle.
func (h *Handle) Pause() {
	h.pause.Store(true)
}

// Continue is continues the handle.
func (h *Handle) Continue() {
	h.pause.Store(false)
}
