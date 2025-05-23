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

package printer

import (
	"io"

	"github.com/kwok-ci/kectl/pkg/client"
	"github.com/kwok-ci/kectl/pkg/encoding"
	"github.com/kwok-ci/kectl/pkg/scheme"
)

type jsonPrinter struct {
	w io.Writer
}

func (p *jsonPrinter) Print(kv *client.KeyValue) error {
	value := kv.Value
	inMediaType, err := encoding.DetectMediaType(value)
	if err != nil {
		return err
	}
	_, data, err := encoding.Convert(scheme.Codecs, inMediaType, encoding.JSONMediaType, value)
	if err != nil {
		return err
	}
	_, err = p.w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
