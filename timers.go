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

// Task represents a task on the timer channel
type Task struct {
	timer    *time.Timer
	interval bool
	duration time.Duration
	function otto.FunctionCall
}

func quit(ctx *Orbit) {
	ctx.Interrupt = make(chan func(), 1)
	if ctx.timeout > 0 {
		go func() {
			time.Sleep(ctx.timeout * time.Millisecond)
			ctx.Interrupt <- func() {}
		}()
	}
}

func wait(ctx *Orbit) {

	for {

		select {
		case <-ctx.Interrupt:
			for timer := range ctx.timers {
				delete(ctx.timers, timer)
			}
		case timer := <-ctx.loop:
			var arguments []interface{}
			if len(timer.function.ArgumentList) > 2 {
				tmp := timer.function.ArgumentList[2:]
				arguments = make([]interface{}, 2+len(tmp))
				for i, value := range tmp {
					arguments[i+2] = value
				}
			} else {
				arguments = make([]interface{}, 1)
			}
			arguments[0] = timer.function.ArgumentList[0]
			_, err := ctx.Call(`Function.call.call`, nil, arguments...)
			if err != nil {
				for _, timer := range ctx.timers {
					timer.timer.Stop()
					delete(ctx.timers, timer)
				}
				for _, e := range faults {
					if e.when == "fail" {
						e.what(ctx, err)
					}
				}
			}
			if timer.interval {
				timer.timer.Reset(timer.duration)
			} else {
				delete(ctx.timers, timer)
			}
		default:
			// Escape valve!
		}

		if len(ctx.timers) == 0 {
			break
		}

	}

}

func init() {

	OnInit(func(ctx *Orbit) {

		ctx.Set("setTimeout", func(call otto.FunctionCall) otto.Value {

			delay, _ := call.Argument(1).ToInteger()
			if delay <= 0 {
				delay = 1
			}

			timer := &Task{
				function: call,
				interval: false,
				duration: time.Duration(delay) * time.Millisecond,
			}

			ctx.timers[timer] = timer

			timer.timer = time.AfterFunc(timer.duration, func() {
				ctx.loop <- timer
			})

			val, err := call.Otto.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		ctx.Set("setInterval", func(call otto.FunctionCall) otto.Value {

			delay, _ := call.Argument(1).ToInteger()
			if delay <= 0 {
				delay = 1
			}

			timer := &Task{
				function: call,
				interval: true,
				duration: time.Duration(delay) * time.Millisecond,
			}

			ctx.timers[timer] = timer

			timer.timer = time.AfterFunc(timer.duration, func() {
				ctx.loop <- timer
			})

			val, err := call.Otto.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		ctx.Set("setImmediate", func(call otto.FunctionCall) otto.Value {

			timer := &Task{
				function: call,
				interval: false,
				duration: time.Millisecond,
			}

			ctx.timers[timer] = timer

			timer.timer = time.AfterFunc(timer.duration, func() {
				ctx.loop <- timer
			})

			val, err := call.Otto.ToValue(timer)
			if err != nil {
				panic(err)
			}

			return val

		})

		ctx.Set("clearTimeout", func(call otto.FunctionCall) otto.Value {
			timer, _ := call.Argument(0).Export()
			if timer, ok := timer.(*Task); ok {
				timer.timer.Stop()
				delete(ctx.timers, timer)
			}
			return otto.UndefinedValue()
		})

		ctx.Set("clearImmediate", func(call otto.FunctionCall) otto.Value {
			timer, _ := call.Argument(0).Export()
			if timer, ok := timer.(*Task); ok {
				timer.timer.Stop()
				delete(ctx.timers, timer)
			}
			return otto.UndefinedValue()
		})

	})

}
