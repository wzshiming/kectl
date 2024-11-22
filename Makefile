
.PHONY: generate
generate: \
	pkg/wellknown/resources.go \
	pkg/scheme/scheme.go \
	pkg/old/scheme/scheme.go

pkg/scheme/scheme.go: ./hack/gen_scheme.sh go.mod
	go mod vendor
	-rm ./pkg/scheme/scheme.go
	./hack/gen_scheme.sh > ./pkg/scheme/scheme.go

pkg/old/api:
	./hack/clone_old_apis.sh 30

pkg/old/scheme/scheme.go: ./hack/gen_old_scheme.sh pkg/old/apis go.mod
	-rm ./pkg/old/scheme/scheme.go
	./hack/gen_old_scheme.sh > ./pkg/old/scheme/scheme.go

pkg/wellknown/resources.go: ./hack/gen_wellknown_resources go.mod
	KWOK_KUBE_VERSION=1.30.3 kwokctl create cluster --name kectl-wellknown \
		--kube-apiserver-insecure-port 8080 \
		--runtime binary \
		--disable-kube-controller-manager \
		--disable-kube-scheduler \
		--kubeconfig ./kectl-wellknown.kubeconfig
	-rm pkg/wellknown/resources.go
	go run ./hack/gen_wellknown_resources ./kectl-wellknown.kubeconfig > pkg/wellknown/resources.go
	kwokctl delete cluster --name kectl-wellknown \
		--kubeconfig ./kectl-wellknown.kubeconfig
	rm ./kectl-wellknown.kubeconfig
