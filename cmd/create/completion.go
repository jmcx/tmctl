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

package create

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tmctl/cmd/brokers"
	"github.com/triggermesh/tmctl/pkg/completion"
	"github.com/triggermesh/tmctl/pkg/triggermesh/crd"
)

func (o *CreateOptions) sourcesCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		sources, err := crd.ListSources(o.CRD)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return sources, cobra.ShellCompDirectiveNoFileComp
	}
	if args[len(args)-1] == "--broker" {
		list, err := brokers.List(o.ConfigBase, "")
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		return list, cobra.ShellCompDirectiveNoFileComp
	}
	if toComplete == "--name" {
		return []string{toComplete}, cobra.ShellCompDirectiveNoFileComp
	}
	if strings.HasPrefix(args[len(args)-1], "--") {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	prefix := ""
	toComplete = strings.TrimLeft(toComplete, "-")
	var properties map[string]crd.Property

	if !strings.Contains(toComplete, ".") {
		_, properties = completion.SpecFromCRD(args[0]+"source", o.CRD)
		if property, exists := properties[toComplete]; exists {
			if property.Typ == "object" {
				return []string{"--" + toComplete + "."}, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
			}
			return []string{"--" + toComplete}, cobra.ShellCompDirectiveNoFileComp
		}
	} else {
		path := strings.Split(toComplete, ".")
		exists, nestedProperties := completion.SpecFromCRD(args[0]+"source", o.CRD, path...)
		if len(nestedProperties) != 0 {
			prefix = toComplete
			if !strings.HasSuffix(prefix, ".") && prefix != "--" {
				prefix += "."
			}
			properties = nestedProperties
		} else if exists {
			return []string{"--" + toComplete}, cobra.ShellCompDirectiveNoFileComp
		} else {
			_, properties = completion.SpecFromCRD(args[0]+"source", o.CRD, path[:len(path)-1]...)
			prefix = strings.Join(path[:len(path)-1], ".") + "."
		}
	}

	var spec []string
	for name, property := range properties {
		attr := property.Typ
		if property.Required {
			attr = fmt.Sprintf("required,%s", attr)
		}
		name = prefix + name
		spec = append(spec, fmt.Sprintf("--%s\t(%s) %s", name, attr, property.Description))
	}
	return append(spec, "--name\tOptional component name."), cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

func (o *CreateOptions) targetsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		sources, err := crd.ListTargets(o.CRD)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return sources, cobra.ShellCompDirectiveNoFileComp
	}

	if toComplete == "--source" ||
		toComplete == "--eventTypes" ||
		toComplete == "--name" {
		return []string{toComplete}, cobra.ShellCompDirectiveNoFileComp
	}
	manifestPath := path.Join(o.ConfigBase, o.Context, manifestFile)
	switch args[len(args)-1] {
	case "--source":
		return completion.ListSources(manifestPath), cobra.ShellCompDirectiveNoFileComp
	case "--eventTypes":
		return completion.ListEventTypes(manifestPath, o.CRD), cobra.ShellCompDirectiveNoFileComp
	case "--broker":
		list, err := brokers.List(o.ConfigBase, "")
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		return list, cobra.ShellCompDirectiveNoFileComp
	}
	if strings.HasPrefix(args[len(args)-1], "--") {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	prefix := ""
	toComplete = strings.TrimLeft(toComplete, "-")
	var properties map[string]crd.Property

	if !strings.Contains(toComplete, ".") {
		_, properties = completion.SpecFromCRD(args[0]+"target", o.CRD)
		if property, exists := properties[toComplete]; exists {
			if property.Typ == "object" {
				return []string{"--" + toComplete + "."}, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
			}
			return []string{"--" + toComplete}, cobra.ShellCompDirectiveNoFileComp
		}
	} else {
		path := strings.Split(toComplete, ".")
		exists, nestedProperties := completion.SpecFromCRD(args[0]+"target", o.CRD, path...)
		if len(nestedProperties) != 0 {
			prefix = toComplete
			if !strings.HasSuffix(prefix, ".") && prefix != "--" {
				prefix += "."
			}
			properties = nestedProperties
		} else if exists {
			return []string{"--" + toComplete}, cobra.ShellCompDirectiveNoFileComp
		} else {
			_, properties = completion.SpecFromCRD(args[0]+"target", o.CRD, path[:len(path)-1]...)
			prefix = strings.Join(path[:len(path)-1], ".") + "."
		}
	}

	var spec []string
	for name, property := range properties {
		attr := property.Typ
		if property.Required {
			attr = fmt.Sprintf("required,%s", attr)
		}
		name = prefix + name
		spec = append(spec, fmt.Sprintf("--%s\t(%s) %s", name, attr, property.Description))
	}
	return append(spec,
		"--source\tEvent source name.",
		"--eventTypes\tEvent types filter.",
		"--name\tOptional component name.",
	), cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}