/*
Copyright 2020 The Kubernetes Authors.

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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/zeitgeist/dependency"
)

func addValidate(topLevel *cobra.Command) {
	vo := rootOpts

	cmd := &cobra.Command{
		Use:           "validate",
		Short:         "Check dependencies locally and against upstream versions",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(*cobra.Command, []string) error {
			return vo.setAndValidate()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runValidate(vo)
		},
	}

	topLevel.AddCommand(cmd)
}

// runValidate is the function invoked by 'addValidate', responsible for
// validating dependencies in a specified configuration file.
func runValidate(opts *options) error {
	var (
		client dependency.Client
		err    error
	)
	if opts.localOnly {
		client, err = dependency.NewLocalClient()
	} else {
		client, err = dependency.NewRemoteClient()
	}
	if err != nil {
		return fmt.Errorf("constructing client: %w", err)
	}

	if err := client.LocalCheck(opts.configFile, opts.basePath); err != nil {
		return fmt.Errorf("checking local dependencies: %w", err)
	}

	if !opts.localOnly {
		updates, err := client.RemoteCheck(opts.configFile)
		if err != nil {
			return fmt.Errorf("checking remote dependencies: %w", err)
		}

		for _, update := range updates {
			fmt.Println(update)
		}
	}

	return nil
}
