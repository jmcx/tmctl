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

package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/triggermesh/tmcli/pkg/openapi"
	"github.com/triggermesh/tmcli/pkg/triggermesh/crd"
)

const (
	labelKey = "triggermesh.io/context"
)

type Object struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   Metadata               `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
}

type Metadata struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels"`
}

func CreateObject(resource, name, broker, crdFile string, spec map[string]interface{}) (*Object, error) {
	crdObject, err := crd.GetResourceCRD(resource, crdFile)
	if err != nil {
		return nil, fmt.Errorf("CRD schema not found: %w", err)
	}
	schema, version, err := getObjectCRD(crdObject)
	if err != nil {
		return nil, fmt.Errorf("object schema: %w", err)
	}
	if spec, err = schema.Process(spec); err != nil {
		return nil, fmt.Errorf("spec processing: %w", err)
	}
	if err := schema.Validate(spec); err != nil {
		return nil, fmt.Errorf("CR validation: %w", err)
	}
	return &Object{
		APIVersion: fmt.Sprintf("%s/%s", crdObject.Spec.Group, version),
		Kind:       crdObject.Spec.Names.Kind,
		Metadata: Metadata{
			Name: name,
			Labels: map[string]string{
				labelKey: broker,
			},
		},
		Spec: spec,
	}, nil
}

func CreateUnstructured(resource, name, broker, crdFile string, spec map[string]interface{}) (*unstructured.Unstructured, error) {
	crdObject, err := crd.GetResourceCRD(resource, crdFile)
	if err != nil {
		return nil, fmt.Errorf("CRD schema not found: %w", err)
	}
	schema, version, err := getObjectCRD(crdObject)
	if err != nil {
		return nil, fmt.Errorf("object schema: %w", err)
	}
	if spec, err = schema.Process(spec); err != nil {
		return nil, fmt.Errorf("spec processing: %w", err)
	}
	if err := schema.Validate(spec); err != nil {
		return nil, fmt.Errorf("CR validation: %w", err)
	}
	u := &unstructured.Unstructured{}
	u.SetAPIVersion(fmt.Sprintf("%s/%s", crdObject.Spec.Group, version))
	u.SetKind(crdObject.Spec.Names.Kind)
	u.SetName(name)
	u.SetLabels(map[string]string{labelKey: broker})
	return u, unstructured.SetNestedField(u.Object, spec, "spec")
}

func getObjectCRD(crdObject crd.CRD) (*openapi.Schema, string, error) {
	for _, v := range crdObject.Spec.Versions {
		if v.Served {
			schema, err := openapi.GetSchema(v.Schema.OpenAPIV3Schema.Properties.Spec)
			if err != nil {
				return nil, "", fmt.Errorf("CRD schema: %w", err)
			}
			return schema, v.Name, nil
		}
	}
	return nil, "", fmt.Errorf("CRD schema not found")
}