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

type Task interface {
	Startup(*Orbit)
	Cleanup(*Orbit)
	Execute(*Orbit) error
}

func (ctx *Orbit) Push(t Task) {
	ctx.lock.Lock()
	ctx.tasks[t] = t
	t.Startup(ctx)
	ctx.lock.Unlock()
}

func (ctx *Orbit) Pull(t Task) {
	ctx.lock.Lock()
	delete(ctx.tasks, t)
	t.Cleanup(ctx)
	ctx.lock.Unlock()
}

func (ctx *Orbit) Next(t Task) {
	ctx.queue <- t
}

func (ctx *Orbit) loop() (err error) {

	for {

		select {

		default:

		case err := <-ctx.quit:

			panic(err)

		case task := <-ctx.queue:

			if err := task.Execute(ctx); err != nil {
				panic(err)
			}

		}

		if len(ctx.tasks) == 0 {
			break
		}

	}

	if ctx.timer != nil {
		ctx.timer.Stop()
	}

	return

}
