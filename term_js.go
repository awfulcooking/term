// Copyright 2021 mooff@cyberspace.baby. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js
// +build js

package term

import (
	"errors"
	"fmt"
	"runtime"
	"syscall/js"
)

// Passthrough for JS environments running e.g. xterm.js
//
// Create a key on the global object (globalThis, window) pointing to
// an object with the following functions:
//     isTerminal(fd)     => bool
//     getSize(fd)        => {width, height} object
//     getState(fd)       => custom, opaque state object
//     makeRaw(fd)        => custom, opaque state object
//     restore(fd, state)
//
// Unused functions can be omitted, but their Golang callers will
// indicate this by returning an error.

var jsErrorCtor js.Value

func init() {
	jsErrorCtor = js.Global().Get("Error")
}

func jsBridge(method string, args ...interface{}) (error, js.Value) {
	result := js.Global().Get("golang.org/x/term").Call(method, args...)
	if result.InstanceOf(jsErrorCtor) {
		return js.Error{result}, js.Null()
	}
	return nil, result
}

type state js.Value

func isTerminal(fd int) bool {
	_, result := jsBridge("isTerminal", fd)
	return result.Truthy()
}

func makeRaw(fd int) (*State, error) {
	if err, newState := jsBridge("makeRaw", fd); err == nil {
		return &State{state(newState)}, nil
	} else {
		return nil, err
	}
}

func getState(fd int) (*State, error) {
	if err, newState := jsBridge("getState", fd); err == nil {
		return &State{state(newState)}, nil
	} else {
		return nil, err
	}
}

func restore(fd int, state *State) error {
	err, _ := jsBridge("restore", fd, state.state)
	return err
	// return nil
}

func getSize(fd int) (width, height int, err error) {
	if err, wh := jsBridge("getSize", fd); err != nil {
		return 0, 0, err
	} else if t := wh.Type(); t != js.TypeObject {
		return 0, 0, errors.New("JS getSize(fd) must return an object with width and height properties")
	} else if w, h := wh.Get("width"), wh.Get("height"); w.Type() != js.TypeNumber || h.Type() != js.TypeNumber {
		return 0, 0, errors.New("JS getSize(fd) must return an object with width and height properties")
	} else {
		return w.Int(), h.Int(), nil
	}
}

func readPassword(fd int) ([]byte, error) {
	return nil, fmt.Errorf("terminal: ReadPassword not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}
