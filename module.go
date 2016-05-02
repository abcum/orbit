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
	"github.com/robertkrimen/otto"
)

type module func(*Orbit) (otto.Value, error)

func null() module {
	return func(ctx *Orbit) (val otto.Value, err error) {
		return otto.UndefinedValue(), nil
	}
}

func find(name, path string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		if len(name) == 0 {
			return otto.UndefinedValue(), fmt.Errorf("No module name specified")
		}

		return otto.UndefinedValue(), fmt.Errorf("Module %s was not found", name)

	}
}

func load(code, path string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		data := "(function(module) { var require = module.require; var exports = module.exports;\n" + code + "\n})"

		module, _ := ctx.Object(`({exports: {}})`)
		export, _ := module.Get("exports")

		module.Set("require", func(call otto.FunctionCall) otto.Value {
			arg := call.Argument(0).String()
			val, err := ctx.require(arg, path)
			if err != nil {
				ctx.Call("new Error", nil, err.Error())
			}
			return val
		})

		ret, err := ctx.Call(data, export, module)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		if ret.IsDefined() {
			val = ret
		}

		if ret.IsUndefined() {
			val, err = module.Get("exports")
		}

		return

	}

}

func (ctx *Orbit) require(name, path string) (val otto.Value, err error) {

	// Check loaded modules
	if module, ok := ctx.modules[name]; ok {
		return module, nil
	}

	// Check global modules
	if module, ok := modules[name]; ok {
		return module(ctx)
	}

	ctx.modules[name], err = find(name, path)(ctx)

	return ctx.modules[name], err

}
