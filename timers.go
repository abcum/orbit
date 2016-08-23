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

type callback func(...interface{})

type timer struct {
	timer    *time.Timer
	interval bool
	callback callback
	argument []interface{}
	duration time.Duration
}

func (t *timer) Startup(ctx *Orbit) {
}

func (t *timer) Cleanup(ctx *Orbit) {
	t.timer.Stop()
}

func (t *timer) Execute(ctx *Orbit) (err error) {

	t.callback(t.argument...)

	if t.interval {
		t.timer.Reset(t.duration)
	} else {
		ctx.Pull(t)
	}

	return

}

func init() {

	OnInit(func(ctx *Orbit) {

		ctx.Set("setTimeout", func(call callback, delay int, args ...interface{}) otto.Value {

			if delay <= 0 {
				delay = 1
			}

			timer := &timer{
				callback: call,
				argument: args,
				interval: false,
				duration: time.Duration(delay) * time.Millisecond,
			}

			ctx.Push(timer)

			timer.timer = time.AfterFunc(timer.duration, func() {
				ctx.Next(timer)
			})

			val, err := ctx.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		ctx.Set("setInterval", func(call callback, delay int, args ...interface{}) otto.Value {

			if delay <= 0 {
				delay = 1
			}

			timer := &timer{
				callback: call,
				argument: args,
				interval: true,
				duration: time.Duration(delay) * time.Millisecond,
			}

			ctx.Push(timer)

			timer.timer = time.AfterFunc(timer.duration, func() {
				ctx.Next(timer)
			})

			val, err := ctx.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		ctx.Set("setImmediate", func(call callback, args ...interface{}) otto.Value {

			timer := &timer{
				callback: call,
				argument: args,
				interval: false,
				duration: time.Millisecond,
			}

			ctx.Push(timer)

			timer.timer = time.AfterFunc(timer.duration, func() {
				ctx.Next(timer)
			})

			val, err := ctx.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		ctx.Set("clearTimeout", func(call otto.FunctionCall) otto.Value {
			name, _ := call.Argument(0).Export()
			if task, ok := name.(*timer); ok {
				task.timer.Stop()
				ctx.Pull(task)
			}
			return otto.UndefinedValue()
		})

		ctx.Set("clearInterval", func(call otto.FunctionCall) otto.Value {
			name, _ := call.Argument(0).Export()
			if task, ok := name.(*timer); ok {
				task.timer.Stop()
				ctx.Pull(task)
			}
			return otto.UndefinedValue()
		})

		ctx.Set("clearImmediate", func(call otto.FunctionCall) otto.Value {
			name, _ := call.Argument(0).Export()
			if task, ok := name.(*timer); ok {
				task.timer.Stop()
				ctx.Pull(task)
			}
			return otto.UndefinedValue()
		})

	})

}
