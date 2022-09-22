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
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/triggermesh/tmcli/pkg/triggermesh"
	tmbroker "github.com/triggermesh/tmcli/pkg/triggermesh/broker"
	"github.com/triggermesh/tmcli/pkg/triggermesh/crd"
	"github.com/triggermesh/tmcli/pkg/triggermesh/source"
	"github.com/triggermesh/tmcli/pkg/triggermesh/target"
)

func (o *CreateOptions) NewTargetCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "target <kind> <args>",
		Short:              "TriggerMesh target",
		DisableFlagParsing: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.initializeOptions(cmd)
			if len(args) == 0 {
				sources, err := crd.ListTargets(o.CRD)
				if err != nil {
					return fmt.Errorf("list sources: %w", err)
				}
				fmt.Printf("Available targets:\n---\n%s\n", strings.Join(sources, "\n"))
				return nil
			}
			kind, args, err := parse(args)
			if err != nil {
				return err
			}
			sourceFilter, args := parameterFromArgs("source", args)
			eventTypesFilter, args := parameterFromArgs("eventTypes", args)
			if sourceFilter == "" && eventTypesFilter == "" {
				return fmt.Errorf("\"--source=<kind>\" or \"--eventTypes=<a,b,c>\" is required")
			}
			var eventFilter []string
			if eventTypesFilter != "" {
				eventFilter = strings.Split(eventTypesFilter, ",")
			}
			return o.target(kind, args, sourceFilter, eventFilter)
		},
	}
}

func parameterFromArgs(parameter string, args []string) (string, []string) {
	var value string
	for k := 0; k < len(args); k++ {
		if strings.HasPrefix(args[k], "--"+parameter) {
			if kv := strings.Split(args[k], "="); len(kv) == 2 {
				value = kv[1]
			} else if len(args) > k+1 && !strings.HasPrefix(args[k+1], "--") {
				value = args[k+1]
				k++
			}
			args = append(args[:k-1], args[k+1:]...)
			break
		}
	}
	return value, args
}

func (o *CreateOptions) target(kind string, args []string, sourceKind string, eventTypes []string) error {
	ctx := context.Background()
	configDir := path.Join(o.ConfigBase, o.Context)

	var s triggermesh.Component
	if sourceKind != "" {
		s = source.New(o.CRD, sourceKind, o.Context, o.Version, nil)
		et, err := s.(*source.Source).GetEventTypes()
		if err != nil {
			return fmt.Errorf("source event types: %w", err)
		}
		eventTypes = append(eventTypes, et...)
	}

	t := target.New(o.CRD, kind, o.Context, o.Version, args)
	log.Println("Updating manifest")
	restart, err := triggermesh.Create(ctx, t, path.Join(configDir, manifestFile))
	if err != nil {
		return err
	}
	log.Println("Starting container")
	container, err := triggermesh.Start(ctx, t, restart)
	if err != nil {
		return err
	}

	tr := tmbroker.NewTrigger(fmt.Sprintf("%s-trigger", t.GetName()), o.Context, configDir, eventTypes)
	tr.SetTarget(container.Name, fmt.Sprintf("http://host.docker.internal:%s", container.HostPort()))
	if err := tr.UpdateBrokerConfig(); err != nil {
		return fmt.Errorf("broker config: %w", err)
	}
	if err := tr.UpdateManifest(); err != nil {
		return fmt.Errorf("broker manifest: %w", err)
	}

	sourceMsg := strings.Join(eventTypes, ", ")
	if s != nil {
		sourceMsg = fmt.Sprintf("%s(%s)", s.GetKind(), sourceMsg)
	}
	fmt.Println("---")
	fmt.Printf("Target is subscribed to:\n\t%s\n", sourceMsg)
	fmt.Println("\nNext steps:")
	fmt.Println("\ttmcli watch\t - show events flowing through the broker in the real time")
	fmt.Println("\ttmcli dump\t - dump Kubernetes manifest")
	fmt.Println("---")

	return nil
}
