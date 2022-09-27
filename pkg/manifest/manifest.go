/*
Copyright 2022 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifest

import (
	"io"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"

	"github.com/triggermesh/tmcli/pkg/kubernetes"
)

type Manifest struct {
	Path    string
	Objects []kubernetes.Object
}

func New(path string) *Manifest {
	return &Manifest{
		Path: path,
	}
}

func (m *Manifest) Read() error {
	o, err := parseYAML(m.Path)
	if err != nil {
		return err
	}
	m.Objects = o
	return nil
}

func (m *Manifest) Write() error {
	f, err := os.OpenFile(m.Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, object := range m.Objects {
		body, err := yaml.Marshal(object)
		if err != nil {
			return err
		}
		if _, err := f.WriteString("---\n"); err != nil {
			return err
		}
		if _, err := f.Write(body); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manifest) Add(object kubernetes.Object) (bool, error) {
	for i, o := range m.Objects {
		if matchObjects(object, o) {
			if !reflect.DeepEqual(o, object) {
				m.Objects[i] = object
				return true, nil
			}
			return false, nil
		}
	}
	m.Objects = append(m.Objects, object)
	return true, nil
}

func parseYAML(path string) ([]kubernetes.Object, error) {
	f, err := os.Open(path)
	if err != nil {
		return []kubernetes.Object{}, err
	}

	fstat, err := f.Stat()
	if err != nil {
		return []kubernetes.Object{}, err
	}

	if !fstat.IsDir() {
		return readFile(f)
	}

	entries, err := f.ReadDir(-1)
	if err != nil {
		return []kubernetes.Object{}, err
	}

	var result []kubernetes.Object
	for _, d := range entries {
		if d.IsDir() {
			// Skip directories
			continue
		}
		objects, err := parseYAML(path + "/" + d.Name())
		if err != nil {
			return []kubernetes.Object{}, err
		}
		result = append(result, objects...)
	}
	return result, nil
}

func readFile(file io.Reader) ([]kubernetes.Object, error) {
	var result []kubernetes.Object
	decoder := yaml.NewDecoder(file)
	for {
		o := new(kubernetes.Object)
		err := decoder.Decode(&o)
		if o == nil {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return []kubernetes.Object{}, err
		}
		result = append(result, *o)
	}
	return result, nil
}

func matchObjects(a, b kubernetes.Object) bool {
	return (a.APIVersion == b.APIVersion) &&
		(a.Kind == b.Kind) &&
		(a.Metadata.Name == b.Metadata.Name)
}