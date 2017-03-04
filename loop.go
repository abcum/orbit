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

// Task is a job that can be added to the asynchronous queue.
type Task interface {
	// Startup is called when the task is pushed onto the queue.
	Startup(*Orbit)
	// Cleanup is called when the task is pulled from the queue.
	Cleanup(*Orbit)
	// Execute is called when the task is being called from the run loop.
	Execute(*Orbit) error
}

// Push pushes an asynchronous task onto the queue, ensuring that the script does not finish before the task is complete.
func (orb *Orbit) Push(t Task) {
	orb.lock.Lock()
	orb.tasks[t] = t
	t.Startup(orb)
	orb.lock.Unlock()
}

// Pull removes an asynchronous task from the queue, cleaning up any context data. If all asynchronous events are completed, the script will finish.
func (orb *Orbit) Pull(t Task) {
	orb.lock.Lock()
	delete(orb.tasks, t)
	t.Cleanup(orb)
	orb.lock.Unlock()
}

// Next signals to the run loop that an asynchronous task is ready to be run.
func (orb *Orbit) Next(t Task) {
	orb.queue <- t
}

func (orb *Orbit) loop() (err error) {

	for {

		select {

		default:

		case err := <-orb.quit:

			panic(err)

		case task := <-orb.queue:

			if err := task.Execute(orb); err != nil {
				panic(err)
			}

		}

		if len(orb.tasks) == 0 {
			break
		}

	}

	if orb.timer != nil {
		orb.timer.Stop()
	}

	return

}
