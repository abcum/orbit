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
	"github.com/robertkrimen/otto"
)

// Orbit is a node.go context.
type Orbit struct {
	// Underlying Otto instance.
	*otto.Otto
	// Module outputs are cached for future use.
	modules map[string]otto.Value
}

var globals = make(map[string]global)
var modules = make(map[string]module)

// Add adds a module to the runtime
func Add(name string, item interface{}) {
	switch what := item.(type) {
	case string:
		addSource(name, what)
	case func(*Orbit) (otto.Value, error):
		addModule(name, what)
	}
}

// Set sets a global variable in the runtime
func Set(name string, item interface{}) {
	globals[name] = item
}

func new() *Orbit {
	return &Orbit{otto.New(), make(map[string]otto.Value)}
}

// Run executes some code
func Run(code string) (otto.Value, error) {
	return new().init().main(code)
}

func (ctx *Orbit) init() *Orbit {
	for k, v := range globals {
		ctx.Set(k, v)
	}
	return ctx
}

func (ctx *Orbit) main(code string) (val otto.Value, err error) {
	return load(code, ".")(ctx)
}
