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

type timer struct {
	timer    *time.Timer
	interval bool
	callback otto.Value
	argument []interface{}
	duration time.Duration
}

func (t *timer) Startup(orb *Orbit) {
}

func (t *timer) Cleanup(orb *Orbit) {
	t.timer.Stop()
}

func (t *timer) Execute(orb *Orbit) (err error) {

	t.callback.Call(t.callback, t.argument...)

	if t.interval {
		t.timer.Reset(t.duration)
	} else {
		orb.Pull(t)
	}

	return

}

func init() {

	OnInit(func(orb *Orbit) {

		orb.Set("setTimeout", func(call otto.Value, delay int, args ...interface{}) otto.Value {

			if delay <= 0 {
				delay = 1
			}

			timer := &timer{
				callback: call,
				argument: args,
				interval: false,
				duration: time.Duration(delay) * time.Millisecond,
			}

			orb.Push(timer)

			timer.timer = time.AfterFunc(timer.duration, func() {
				orb.Next(timer)
			})

			val, err := orb.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		orb.Set("setInterval", func(call otto.Value, delay int, args ...interface{}) otto.Value {

			if delay <= 0 {
				delay = 1
			}

			timer := &timer{
				callback: call,
				argument: args,
				interval: true,
				duration: time.Duration(delay) * time.Millisecond,
			}

			orb.Push(timer)

			timer.timer = time.AfterFunc(timer.duration, func() {
				orb.Next(timer)
			})

			val, err := orb.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		orb.Set("setImmediate", func(call otto.Value, args ...interface{}) otto.Value {

			timer := &timer{
				callback: call,
				argument: args,
				interval: false,
				duration: time.Millisecond,
			}

			orb.Push(timer)

			timer.timer = time.AfterFunc(timer.duration, func() {
				orb.Next(timer)
			})

			val, err := orb.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		orb.Set("clearTimeout", func(call otto.FunctionCall) otto.Value {
			name, _ := call.Argument(0).Export()
			if task, ok := name.(*timer); ok {
				task.timer.Stop()
				orb.Pull(task)
			}
			return otto.UndefinedValue()
		})

		orb.Set("clearInterval", func(call otto.FunctionCall) otto.Value {
			name, _ := call.Argument(0).Export()
			if task, ok := name.(*timer); ok {
				task.timer.Stop()
				orb.Pull(task)
			}
			return otto.UndefinedValue()
		})

		orb.Set("clearImmediate", func(call otto.FunctionCall) otto.Value {
			name, _ := call.Argument(0).Export()
			if task, ok := name.(*timer); ok {
				task.timer.Stop()
				orb.Pull(task)
			}
			return otto.UndefinedValue()
		})

	})

}
