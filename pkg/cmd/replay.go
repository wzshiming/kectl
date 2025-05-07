/*
Copyright 2024 The Kubernetes Authors.

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

package cmd

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/kwok-ci/kectl/pkg/client"
	"github.com/kwok-ci/kectl/pkg/scheme"
	"github.com/kwok-ci/kectl/pkg/snapshot"
	"github.com/kwok-ci/kectl/pkg/snapshot/handle"
	"github.com/kwok-ci/kectl/pkg/utils/compress"
	"github.com/kwok-ci/kectl/pkg/utils/yaml"
	"github.com/spf13/cobra"
)

type replayFlagpole struct {
	Prefix   string
	Path     string
	Snapshot bool
}

func newCtlReplayCommand() *cobra.Command {
	flags := &replayFlagpole{}

	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "replay",
		Short: "Replay the recording to the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			etcdclient, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			err = replayCommand(cmd.Context(), etcdclient, flags)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Prefix, "prefix", "/registry", "prefix to prepend to the resource")
	cmd.Flags().StringVar(&flags.Path, "path", "", "Path to the recording")
	cmd.Flags().BoolVar(&flags.Snapshot, "snapshot", false, "Only save the snapshot")
	return cmd
}

func replayCommand(ctx context.Context, etcdclient client.Client, flags *replayFlagpole) error {
	loader, err := snapshot.NewLoader(snapshot.LoadConfig{
		Client: etcdclient,
		Prefix: flags.Prefix,
		Codecs: scheme.Codecs,
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}

	f, err := os.Open(flags.Path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	rc, err := compress.Decompress(flags.Path, f)
	if err != nil {
		return err
	}
	defer func() {
		_ = rc.Close()
	}()

	var reader io.Reader = rc

	startTime := time.Now()
	reader = handle.NewReadHook(reader, func(bytes []byte) []byte {
		return handle.RevertTimeFromRelative(startTime, bytes)
	})

	decoder := yaml.NewDecoder(reader)

	err = loader.Load(ctx, decoder)
	if err != nil {
		return err
	}

	if flags.Snapshot {
		return nil
	}

	if isTerminal() {
		cancel := loader.AllowHandle(ctx)
		defer cancel()
	}
	err = loader.Replay(ctx, decoder)
	if err != nil {
		return err
	}

	return nil
}
