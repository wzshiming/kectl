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
	"log/slog"
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

type recordFlagpole struct {
	Prefix   string
	Path     string
	Snapshot bool
}

func newCtlRecordCommand() *cobra.Command {
	flags := &recordFlagpole{}

	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "record",
		Short: "Record the recording from the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			etcdclient, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			err = recordCommand(cmd.Context(), etcdclient, flags)
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

func recordCommand(ctx context.Context, etcdclient client.Client, flags *recordFlagpole) error {
	saver, err := snapshot.NewSaver(snapshot.SaveConfig{
		Client: etcdclient,
		Prefix: flags.Prefix,
		Codecs: scheme.Codecs,
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}

	f, err := os.OpenFile(flags.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	wc := compress.Compress(flags.Path, f)
	defer func() {
		_ = wc.Close()
	}()

	var writer io.Writer = wc

	startTime := time.Now()
	writer = handle.NewWriteHook(writer, func(bytes []byte) []byte {
		return handle.ReplaceTimeToRelative(startTime, bytes)
	})

	encoder := yaml.NewEncoder(writer)

	if flags.Snapshot {
		slog.Info("Saving snapshot")
	} else {
		slog.Info("Saving snapshot and recording")
	}

	err = saver.Save(ctx, encoder)
	if err != nil {
		return err
	}

	if flags.Snapshot {
		slog.Info("Saved snapshot")
		return nil
	}

	slog.Info("Recording")
	slog.Info("Press Ctrl+C to stop recording resources")

	err = saver.Record(ctx, encoder)
	if err != nil {
		return err
	}

	return nil
}
