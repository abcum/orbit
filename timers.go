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

func (ctx *Orbit) tick() {
	if ctx.timeout > 0 {
		ctx.timer = time.AfterFunc(ctx.timeout, func() {
			ctx.forcequit <- func() {}
			ctx.Interrupt <- func() {}
		})
	}
}

func (ctx *Orbit) wait() (err error) {

	for len(ctx.timers) > 0 {

		select {

		case <-ctx.Interrupt:

			ctx.timer.Stop()

			for timer := range ctx.timers {
				timer.timer.Stop()
				delete(ctx.timers, timer)
			}

			return nil

		case <-ctx.forcequit:

			ctx.timer.Stop()

			for _, e := range fails {
				go e(ctx, err)
			}

			for timer := range ctx.timers {
				timer.timer.Stop()
				delete(ctx.timers, timer)
			}

			return fmt.Errorf("Script timed out")

		case timer := <-ctx.loop:

			args := make([]interface{}, len(timer.function.ArgumentList))
			for i, v := range timer.function.ArgumentList {
				if i != 1 {
					args[i] = v
				}
			}

			if _, err := ctx.Call(`Function.call.call`, nil, args...); err != nil {

				ctx.timer.Stop()

				for _, e := range fails {
					go e(ctx, err)
				}

				for _, timer := range ctx.timers {
					timer.timer.Stop()
					delete(ctx.timers, timer)
				}

				return err

			}

			if timer.interval {
				timer.timer.Reset(timer.duration)
			} else {
				delete(ctx.timers, timer)
			}

		}

	}

	ctx.timer.Stop()

	return

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
