// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orbit

import (
	"time"

	"github.com/robertkrimen/otto"
)

// Orbit is a node.go context.
type Orbit struct {
	// Underlying Otto instance.
	*otto.Otto
	// External runtime variables.
	Vars map[string]interface{}
	// quit
	quit chan error
	// Loop runs pending timers
	loop chan *task
	// Timeout timer
	timer *time.Timer
	// Runtime timers
	timers map[*task]*task
	// Timeout sets a timeout
	timeout time.Duration
	// Module outputs are cached for future use.
	modules map[string]otto.Value
}

type (
	// Module is a javascript module
	module func(*Orbit) (otto.Value, error)
)

var (
	// Event groups
	inits []func(*Orbit)
	exits []func(*Orbit)
	fails []func(*Orbit, error)
	// Finder loads files
	finder func(*Orbit, []string) (interface{}, string, error)
	// Modules stores registered packages
	modules = make(map[string]module)
)

// Add adds a module to the runtime
func Add(name string, item interface{}) {
	switch what := item.(type) {
	case string:
		addSource(name, what)
	case []byte:
		addSource(name, what)
	case func(*Orbit) (otto.Value, error):
		addModule(name, what)
	}
}

func Find(ctx *Orbit, files []string) (interface{}, string, error) {
	return finder(ctx, files)
}

// OnInit registers a callback for when the program starts up
func OnInit(call func(*Orbit)) {
	inits = append(inits, call)
}

// OnExit registers a callback for when the program shuts down
func OnExit(call func(*Orbit)) {
	exits = append(exits, call)
}

// OnFail registers a callback for when the program encounters and error
func OnFail(call func(*Orbit, error)) {
	fails = append(fails, call)
}

// OnFile registers a callback for finding required files
func OnFile(call func(*Orbit, []string) (interface{}, string, error)) {
	finder = call
}

// New creates a new Orbit runtime
func New(timeout time.Duration) *Orbit {

	orbit := &Orbit{
		Otto:    otto.New(),
		Vars:    make(map[string]interface{}),
		quit:    make(chan error),
		loop:    make(chan *task),
		timers:  make(map[*task]*task),
		modules: make(map[string]otto.Value),
		timeout: timeout * time.Millisecond,
	}

	orbit.Interrupt = make(chan func(), 1)

	return orbit

}

// Def sets a global variable in the runtime
func (ctx *Orbit) Def(name string, item interface{}) {
	obj, _ := ctx.Get("global")
	obj.Object().Set(name, item)
	ctx.Set(name, item)
}

// Run executes some code. Code may be a string or a byte slice.
func (ctx *Orbit) Run(name string, code interface{}) (err error) {

	defer func() {

		var ok bool

		if err, ok = recover().(error); ok {
			ctx.fail(err)
		}

		ctx.exit()

		if ctx.timer != nil {
			ctx.timer.Stop()
		}

		for timer := range ctx.timers {
			timer.timer.Stop()
			delete(ctx.timers, timer)
		}

	}()

	ctx.SetStackDepthLimit(20000)

	ctx.tick()

	ctx.init()

	// Run main code
	_, err = main(code, name)(ctx)
	if err != nil {
		panic(err)
	}

	// Wait for timers
	err = ctx.wait()
	if err != nil {
		panic(err)
	}

	return

}

func (ctx *Orbit) Quit(err error) {
	panic(err)
}

func (ctx *Orbit) init() {
	for _, e := range inits {
		e(ctx)
	}
}

func (ctx *Orbit) exit() {
	for _, e := range exits {
		e(ctx)
	}
}

func (ctx *Orbit) fail(err error) {
	for _, e := range fails {
		e(ctx, err)
	}
}
