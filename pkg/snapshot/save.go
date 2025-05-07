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

package snapshot

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kwok-ci/kectl/pkg/client"
	"github.com/kwok-ci/kectl/pkg/encoding"
	"github.com/kwok-ci/kectl/pkg/snapshot/handle"
	"github.com/kwok-ci/kectl/pkg/utils/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/clock"
)

// SaveConfig is the combination of the impersonation config
// and the PagerConfig.
type SaveConfig struct {
	Client client.Client
	Scheme *runtime.Scheme
	Codecs serializer.CodecFactory
	Prefix string
}

// Saver is a snapshot saver.
type Saver struct {
	SaveConfig

	rev      int64
	track    map[handle.ObjectRef]json.RawMessage
	baseTime time.Time
	clock    clock.PassiveClock
}

// NewSaver creates a new snapshot saver.
func NewSaver(saveConfig SaveConfig) (*Saver, error) {

	return &Saver{
		SaveConfig: saveConfig,
		track:      map[handle.ObjectRef]json.RawMessage{},
		clock:      clock.RealClock{},
	}, nil
}

func (s *Saver) save(encoder *yaml.Encoder, kv *client.KeyValue) error {
	value := kv.Value
	if value == nil {
		value = kv.PrevValue
	}

	inMediaType, err := encoding.DetectMediaType(value)
	if err != nil {
		return err
	}
	_, data, err := encoding.Convert(s.Codecs, inMediaType, encoding.JSONMediaType, value)
	if err != nil {
		return err
	}

	obj := &unstructured.Unstructured{}
	err = obj.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	if obj.GetName() == "" {
		return nil
	}

	err = encoder.Encode(obj)
	if err != nil {
		return err
	}

	s.track[handle.KObj(obj)] = data
	return nil
}

// Save saves the snapshot of cluster
func (s *Saver) Save(ctx context.Context, encoder *yaml.Encoder) error {
	rev, err := s.Client.Get(ctx, s.Prefix,
		client.WithResponse(func(kv *client.KeyValue) error {
			return s.save(encoder, kv)
		}),
	)
	if err != nil {
		return err
	}

	s.rev = rev
	return nil
}

func (s *Saver) buildResourcePatch(kv *client.KeyValue) (*handle.ResourcePatch, error) {
	lastValue := kv.Value
	if lastValue == nil {
		lastValue = kv.PrevValue
	}
	inMediaType, err := encoding.DetectMediaType(lastValue)
	if err != nil {
		return nil, err
	}
	_, data, err := encoding.Convert(s.Codecs, inMediaType, encoding.JSONMediaType, lastValue)
	if err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	err = obj.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	if obj.GetName() == "" {
		return nil, nil
	}

	gvk := obj.GroupVersionKind()

	// Create GVR from GVK using scheme
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s", // Pluralize kind
	}

	rp := handle.ResourcePatch{}
	rp.TypeMeta = handle.ResourcePatchType
	rp.SetTargetGroupVersionResource(gvr)
	rp.SetTargetName(obj.GetName(), obj.GetNamespace())

	// base time is the time of the first object
	if s.baseTime.IsZero() {
		now := s.clock.Now()
		s.baseTime = now
	} else {
		now := s.clock.Now()
		rp.SetDuration(now.Sub(s.baseTime))
	}

	switch {
	case kv.Value != nil:
		// Get patch meta directly from scheme
		patchMeta, err := PatchMetaFromStruct(gvk)
		if err != nil {
			return nil, err
		}

		err = rp.SetContent(obj, s.track, patchMeta)
		if err != nil {
			return nil, err
		}
	default:
		rp.SetDelete(obj, s.track)
	}

	return &rp, nil
}

// Record records the snapshot of cluster.
func (s *Saver) Record(ctx context.Context, encoder *yaml.Encoder) error {
	return s.Client.Watch(ctx, s.Prefix,
		client.WithRevision(s.rev),
		client.WithResponse(func(kv *client.KeyValue) error {
			rp, err := s.buildResourcePatch(kv)
			if err != nil {
				return err
			}
			if rp == nil {
				return nil
			}

			return encoder.Encode(rp)
		}),
	)
}
