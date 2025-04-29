/*
Copyright 2025 The Kubernetes Authors.

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

package e2e

import (
	"encoding/json"
	"os/exec"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGet(t *testing.T) {
	out, err := exec.Command(kectl,
		"--endpoints", endpoint,
		"get", "services", "kubernetes", "-n", "default", "-o", "json",
	).Output()
	if err != nil {
		t.Fatal(err)
	}
	got := corev1.Service{}

	err = json.Unmarshal(out, &got)
	if err != nil {
		t.Fatal(err)
	}

	if got.ObjectMeta.Name != "kubernetes" {
		t.Errorf("Got service name: %s, expected kubernetes", got.ObjectMeta.Name)
	}

	if got.ObjectMeta.Namespace != "default" {
		t.Errorf("Got service namespace: %s, expected default", got.ObjectMeta.Namespace)
	}

	if got.TypeMeta.Kind != "Service" {
		t.Errorf("Got service type: %s, expected Service", got.TypeMeta.Kind)
	}
}
