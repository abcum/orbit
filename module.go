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

var (
	beg = "(function(module) { var require = module.require; var exports = module.exports; var __dirname = module.__dirname; var __filename = module.__filename;"
	end = "})"
)

func null() module {
	return func(ctx *Orbit) (val otto.Value, err error) {
		return otto.UndefinedValue(), nil
	}
}

func load(name string, fold string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		// Check loaded modules
		if module, ok := ctx.modules[name]; ok {
			return module, nil
		}

		// Check global modules
		if module, ok := modules[name]; ok {
			return module(ctx)
		}

		ctx.modules[name], err = find(name, fold)(ctx)

		return ctx.modules[name], err

	}
}

func find(name string, fold string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		if len(name) == 0 {
			return otto.UndefinedValue(), &Error{
				fmt.Sprintf("Module '%s' not found", name),
				ctx.Context(),
			}
		}

		var files []string

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
				files = append(files, path.Join(fold, name))
			}
			if path.Ext(name) == "" {
				files = append(files, path.Join(fold, name)+".js")
				files = append(files, path.Join(fold, name, "index.js"))
			}
		}

		code, file, err := finder(ctx, files)
		if err != nil {
			return otto.UndefinedValue(), &Error{
				fmt.Sprintf("Module '%s' not found", name),
				ctx.Context(),
			}
		}

		return exec(code, file)(ctx)

	}
}

func main(code interface{}, full string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		fold, file := path.Split(full)

		script := fmt.Sprintf("%s\n%s\n%s", beg, code, end)

		module, _ := ctx.Object(`({})`)

		module.Set("id", full)
		module.Set("loaded", true)
		module.Set("filename", full)
		module.Set("__dirname", fold)
		module.Set("__filename", file)

		module.Set("require", func(call otto.FunctionCall) otto.Value {
			arg := call.Argument(0).String()
			val, err := load(arg, fold)(ctx)
			if err != nil {
				panic(err)
			}
			return val
		})

		sct, err := ctx.Compile(full, script)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		run, err := ctx.Otto.Run(sct)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		ret, err := run.Call(run, module)
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

func exec(code interface{}, full string) module {
	return func(ctx *Orbit) (val otto.Value, err error) {

		fold, file := path.Split(full)

		script := fmt.Sprintf("%s\n%s\n%s", beg, code, end)

		module, _ := ctx.Object(`({})`)

		module.Set("id", full)
		module.Set("loaded", true)
		module.Set("filename", full)
		module.Set("__dirname", fold)
		module.Set("__filename", file)

		module.Set("require", func(call otto.FunctionCall) otto.Value {
			arg := call.Argument(0).String()
			val, err := load(arg, fold)(ctx)
			if err != nil {
				panic(err)
			}
			return val
		})

		sct, err := ctx.Compile(full, script)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		run, err := ctx.Otto.Run(sct)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		ret, err := run.Call(run, module)
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
