
.PHONY: generate
generate: pkg/wellknown/resources.go

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
