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
	"io/ioutil"
	"path/filepath"
	"strings"
)

func addModule(name string, item module) {
	modules[name] = item
}

func addSource(name string, item interface{}) {

	if data, ok := item.([]byte); ok {

		modules[name] = exec(data, name)

	}

	if file, ok := item.(string); ok {

		if strings.Contains(file, "*") {

			files, _ := filepath.Glob(file)

			for i, file := range files {

				extn := filepath.Ext(file)
				full := filepath.Base(file)
				vers := full[0 : len(full)-len(extn)]

				data, err := ioutil.ReadFile(file)
				if err != nil {
					modules[name+"@"+vers] = null()
					return
				}
				modules[name+"@"+vers] = exec(data, file)

				if i == len(files)-1 {
					modules[name] = modules[name+"@"+vers]
					modules[name+"@latest"] = modules[name+"@"+vers]
				}

			}

		} else {

			data, err := ioutil.ReadFile(file)
			if err != nil {
				modules[name] = null()
				return
			}
			modules[name] = exec(data, file)

		}

	}

}
