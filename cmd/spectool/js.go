/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/dop251/goja"
)

type MacroExpander struct {
	JS *goja.Runtime
}

func (m *MacroExpander) init() error {
	m.JS = goja.New()
	env := make(map[string]interface{})
	m.JS.Set("_", env)

	env["log"] = func(x interface{}) interface{} {
		switch vv := x.(type) {
		case goja.Value:
			x = vv.Export()
		}
		bs, err := json.Marshal(&x)
		if err != nil {
			return err
		}
		log.Printf("%s\n", bs)

		return x
	}

	return nil
}

func (m *MacroExpander) load(filename string) error {
	log.Printf("loading %s", filename)

	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	v, err := m.JS.RunScript(filename, string(src))
	if err != nil {
		return err
	}

	if x := v.Export(); x != nil {
		bs, err := json.Marshal(&x)
		if err != nil {
			return err
		}
		log.Printf("%s → %s\n", filename, bs)
	}
	return nil
}

func (m *MacroExpander) loadMacros(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if !strings.HasSuffix(filename, ".js") {
			continue
		}
		if err = m.load(dir + "/" + filename); err != nil {
			return err
		}
	}

	return nil
}

func MacroExpand(x interface{}) (interface{}, error) {

	js, err := json.Marshal(&x)
	if err != nil {
		return nil, err
	}

	m := &MacroExpander{}

	if err := m.init(); err != nil {
		return nil, err
	}

	if err := m.load("driver.js"); err != nil {
		return nil, err
	}

	if err := m.loadMacros("macros"); err != nil {
		return nil, err
	}

	src := fmt.Sprintf("expand(%s)", js)

	v, err := m.JS.RunString(src)
	if err != nil {
		return nil, err
	}

	return v.Export(), err
}
