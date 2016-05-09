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

	"github.com/robertkrimen/otto"
)

func null() module {
	return func(ctx *Orbit) (val otto.Value, err error) {
		return otto.UndefinedValue(), nil
	}
}

func (ctx *Orbit) load(name string, path string) (val otto.Value, err error) {

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

func find(name string, dir string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		if len(name) == 0 {
			return otto.UndefinedValue(), fmt.Errorf("No module name specified")
		}

		var files []string

		name = path.Clean(name)

		if path.IsAbs(name) == true {
			if path.Ext(name) != "" {
				files = append(files, name)
			}
			if path.Ext(name) == "" {
				files = append(files, name+".js")
				files = append(files, path.Join(name, "index.js"))
			}
		}

		if path.IsAbs(name) == false {
			if path.Ext(name) != "" {
				files = append(files, path.Join(dir, name))
			}
			if path.Ext(name) == "" {
				files = append(files, path.Join(dir, name)+".js")
				files = append(files, path.Join(dir, name, "index.js"))
			}
		}

		code, file, err := finder(ctx, files)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		return exec(code, file)(ctx)

	}
}

func main(code interface{}, dir string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		script := fmt.Sprintf("%s\n%s\n%s", "(function(module) { var require = module.require; var exports = module.exports;\n", code, "\n})")

		module, _ := ctx.Object(`({})`)

		module.Set("id", dir)

		module.Set("loaded", true)

		module.Set("filename", dir)

		module.Set("exports", map[string]interface{}{})

		module.Set("require", func(call otto.FunctionCall) otto.Value {
			arg := call.Argument(0).String()
			val, err := ctx.load(arg, dir)
			if err != nil {
				ctx.Call("new Error", nil, err.Error())
			}
			return val
		})

		ret, err := ctx.Call(script, nil, module)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		exp, err := module.Get("exports")
		if err != nil {
			return otto.UndefinedValue(), err
		}

		if exp.IsFunction() {
			val, err = module.Call("exports")
			return
		}

		if exp.IsDefined() {
			val = exp
			return
		}

		if ret.IsDefined() {
			val = ret
			return
		}

		return

	}

}

func exec(code interface{}, dir string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		data := fmt.Sprintf("%s\n%s\n%s", "(function(module) { var require = module.require; var exports = module.exports;\n", code, "\n})")

		module, _ := ctx.Object(`({exports: {}})`)
		export, _ := module.Get("exports")

		module.Set("require", func(call otto.FunctionCall) otto.Value {
			arg := call.Argument(0).String()
			val, err := ctx.load(arg, dir)
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
