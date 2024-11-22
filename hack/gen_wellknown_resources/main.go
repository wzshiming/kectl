/*
Copyright The Kubernetes Authors.

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

package main

import (
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	err := gen()
	if err != nil {
		panic(err)
	}
}

func gen() error {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: os.Args[1]},
		&clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	resourceList, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		return err
	}

	var wellKnown []resource

	for _, list := range resourceList {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			return err
		}
		for _, ar := range list.APIResources {

			names := []string{ar.Name}
			if ar.SingularName != ar.Name {
				names = append(names, ar.SingularName)
			}
			names = append(names, ar.ShortNames...)

			wellKnown = append(wellKnown, resource{
				Names:      names,
				Namespaced: ar.Namespaced,
				Group:      gv.Group,
			})
		}
	}

	sort.Slice(wellKnown, func(i, j int) bool {
		if wellKnown[i].Group == wellKnown[j].Group {
			return wellKnown[i].Names[0] < wellKnown[j].Names[0]
		}
		return wellKnown[i].Group < wellKnown[j].Group
	})

	out := fmt.Sprintf("%#v", wellKnown)
	out = strings.ReplaceAll(out, "[]main.resource{", "[]resource{\n")
	out = strings.ReplaceAll(out, "main.resource{", "{\n")
	out = strings.ReplaceAll(out, " ", "\n")
	out = strings.ReplaceAll(out, "\"}", "\",\n}")
	out = strings.ReplaceAll(out, "{\"", "{\n\"")
	out = strings.ReplaceAll(out, "}}", "},\n}")

	out = `
/*
Copyright The Kubernetes Authors.

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

package wellknown

type resource struct {
	Names      []string
	Namespaced bool
	Group      string
}

// Don't edit this file directly. It is generated by hack/gen_wellknown_resources
var resources = ` + out

	formated, err := format.Source([]byte(out))
	if err != nil {
		return err
	}
	out = strings.TrimSpace(string(formated))
	fmt.Println(out)

	return nil
}

type resource struct {
	Names      []string
	Namespaced bool
	Group      string
}
