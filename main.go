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
	"fmt"
	"path"
	"sync"
	"time"

	"context"

	"github.com/robertkrimen/otto"
)

// Orbit is a node.go context.
type Orbit struct {
	// Underlying Otto instance.
	*otto.Otto
	// Context
	ctx context.Context
	// Lock mutex
	lock sync.RWMutex
	// Quit channel
	quit chan interface{}
	// Loop channel
	queue chan Task
	// Timeout timer
	timer *time.Timer
	// Runtime tasks
	tasks map[Task]Task
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
		quit:    make(chan interface{}, 1),
		queue:   make(chan Task),
		tasks:   make(map[Task]Task),
		modules: make(map[string]otto.Value),
		timeout: timeout * time.Millisecond,
	}

	orbit.Interrupt = make(chan func(), 1)

	return orbit

}

func (orb *Orbit) Context() context.Context {
	if orb.ctx != nil {
		return orb.ctx
	}
	return context.Background()
}

func (orb *Orbit) WithContext(ctx context.Context) *Orbit {
	nrb := new(Orbit)
	*nrb = *orb
	nrb.ctx = ctx
	return nrb
}

// Def sets a global variable in the runtime
func (orb *Orbit) Def(name string, item interface{}) {
	obj, _ := orb.Get("global")
	obj.Object().Set(name, item)
	orb.Set(name, item)
}

// Exec executes some code. Code may be a string or a byte slice.
func (orb *Orbit) Exec(name string, code interface{}) (err error) {

	defer func() {

		var ok bool

		if err, ok = recover().(error); ok {
			orb.fail(err)
		}

		orb.exit()

		if orb.timer != nil {
			orb.timer.Stop()
		}

		for task := range orb.tasks {
			orb.Pull(task)
		}

	}()

	orb.SetStackDepthLimit(20000)

	orb.tick()

	orb.init()

	// Run main code
	_, err = main(code, name)(orb)
	if err != nil {
		panic(err)
	}

	// Wait for timers
	err = orb.loop()
	if err != nil {
		panic(err)
	}

	return

}

// File finds a file relative to the current javascript context.
func (orb *Orbit) File(name string, extn string) (code interface{}, file string, err error) {

	var files []string

	fold, _ := path.Split(orb.Otto.Context().Filename)

	if path.IsAbs(name) == true {
		if path.Ext(name) != "" {
			files = append(files, name)
		}
		if path.Ext(name) == "" {
			files = append(files, name+"."+extn)
		}
	}

	if path.IsAbs(name) == false {
		if path.Ext(name) != "" {
			files = append(files, path.Join(fold, name))
		}
		if path.Ext(name) == "" {
			files = append(files, path.Join(fold, name)+"."+extn)
		}
	}

	if code, file, err = finder(orb, files); err != nil {
		panic(orb.MakeCustomError("Error", fmt.Sprintf("Cannot find file '%s'", name)))
	}

	return code, name, err

}

// Quit exits the current javascript context cleanly, or with an error.
func (orb *Orbit) Quit(err interface{}) {
	orb.quit <- err
}

func (orb *Orbit) init() {
	for _, e := range inits {
		e(orb)
	}
}

func (orb *Orbit) exit() {
	for _, e := range exits {
		e(orb)
	}
}

func (orb *Orbit) fail(err error) {
	for _, e := range fails {
		e(orb, err)
	}
}

func (orb *Orbit) tick() {
	if orb.timeout > 0 {
		orb.timer = time.AfterFunc(orb.timeout, func() {
			err := fmt.Errorf("Script timeout")
			orb.Interrupt <- func() {
				panic(err)
			}
			orb.quit <- err
		})
	}
}
