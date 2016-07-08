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
	// Loop runs pending timers
	loop chan *Task
	// Timers used within javascript
	timers map[*Task]*Task
	// Timeout sets a timeout
	timeout time.Duration
	// Module outputs are cached for future use.
	modules map[string]otto.Value
}

type (
	// Event is used for storing
	event struct {
		when string
		what func(*Orbit)
	}
	// Global is a global variable
	global interface{}
	// Module is a javascript module
	module func(*Orbit) (otto.Value, error)
	// Finder is a package file loader
	lookup func(*Orbit, []string) (interface{}, string, error)
)

var (
	// Finder loads files
	finder lookup
	// Events stores events
	events []event
	// Globals stores global variables
	globals = make(map[string]global)
	// Modules stores registered packages
	modules = make(map[string]module)
)

// Add adds a module to the runtime
func Add(name string, item interface{}) {
	switch what := item.(type) {
	case string:
		addSource(name, what)
	case func(*Orbit) (otto.Value, error):
		addModule(name, what)
	}
}

// OnFile registers a callback for finding required files
func OnFile(call func(*Orbit, []string) (interface{}, string, error)) {
	finder = call
}

// OnInit registers a callback for when the program starts up
func OnInit(call func(*Orbit)) {
	events = append(events, event{
		when: "init",
		what: call,
	})
}

// OnExit registers a callback for when the program shuts down
func OnExit(call func(*Orbit)) {
	events = append(events, event{
		when: "exit",
		what: call,
	})
}

// Run executes some code. Code may be a string or a byte slice.
func Run(code interface{}) (otto.Value, error) {
	return New().Run(code)
}

// New creates a new Orbit runtime
func New() *Orbit {
	return &Orbit{
		Otto:    otto.New(),
		loop:    make(chan *Task),
		timers:  make(map[*Task]*Task),
		modules: make(map[string]otto.Value),
	}
}

// Def sets a global variable in the runtime
func (ctx *Orbit) Def(name string, item interface{}) {
	obj, _ := ctx.Get("global")
	obj.Object().Set(name, item)
	ctx.Set(name, item)
}

// Run executes some code. Code may be a string or a byte slice.
func (ctx *Orbit) Run(code interface{}) (val otto.Value, err error) {

	ctx.SetStackDepthLimit(20000)

	quit(ctx) // Set a timeout

	ctx.Object(`global = {}`)

	for k, v := range globals {
		ctx.Def(k, v)
	}

	for _, e := range events {
		if e.when == "init" {
			e.what(ctx)
		}
	}

	val, err = main(code, ".")(ctx)
	if err != nil {
		return
	}

	wait(ctx) // Wait for timers

	for _, e := range events {
		if e.when == "exit" {
			e.what(ctx)
		}
	}

	return

}
