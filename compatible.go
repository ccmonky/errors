// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright (c) 2015, Dave Cheney <dave@cheney.net>
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.

// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package errors

import (
	"errors"
	"fmt"
	"runtime"

	pkgerrors "github.com/pkg/errors"
)

// standard
var (
	// New is `errors.New`
	New = errors.New

	// Errorf is `fmt.Errorf`
	Errorf = fmt.Errorf
)

// github.com/pkg/errros
var (
	// WithMessage imitate `github.com/pkg/errors.WithMessage` but implemented with `Message` attr
	WithMessage = MessageAttr.With
)

// WithMessage imitate `github.com/pkg/errors.WithMessagef` but implemented with `Message` attr
func WithMessagef(err error, format string, args ...any) error {
	return WithMessage(err, fmt.Sprintf(format, args...))
}

// WithStack imitate `github.com/pkg/errors.WithStack` but implemented with `Stack` attr
func WithStack(err error) error {
	return StackAttr.With(err, callers())
}

// Wrap imitate `github.com/pkg/errors.Wrap` but implemented with `Message` & `Stack` attrs
func Wrap(err error, message string) error {
	return StackAttr.With(WithMessage(err, message), callers())
}

// Wrapf imitate `github.com/pkg/errors.Wrapf` but implemented with `Message` & `Stack` attrs
func Wrapf(err error, format string, args ...interface{}) error {
	return StackAttr.With(WithMessagef(err, format, args...), callers())
}

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := pkgerrors.Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() pkgerrors.StackTrace {
	f := make([]pkgerrors.Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = pkgerrors.Frame((*s)[i])
	}
	return f
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}
