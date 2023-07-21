current_dir = $(shell pwd)
quick-start:
	make dependencies ip=$(ip) cert_manager=$(cert_manager) load_balancer=$(load_balancer)
	make cert domain=$(domain)
	chmod 777 ./examples/certs/key.pem
	make set-up auto_tls=$(auto_tls) development=$(development)
set-up:
	sed -i 's/AUTOTLSPLACEHOLDER/$(auto_tls)/g' ./sev-csc-demo/values.yaml
	@if [ -n "$(domain)" ]; then sed -i 's/DOMAINPLACEHOLDER/$(domain)/g' ./sev-csc-demo/values.yaml; fi
	sed -i 's/DEVELOPMENTPLACEHOLDER/$(development)/g' ./sev-csc-demo/values.yaml
	helm package sev-csc-demo
	helm install sev-csc-demo sev-csc-demo-0.1.0.tgz
cert:
	mkdir -p ./examples/certs
	sed -i 's/DOMAINPLACEHOLDER/$(domain)/g' ./sev-csc-demo/config/san.cnf
	sed -i 's/DOMAINPLACEHOLDER/$(domain)/g' ./sev-csc-demo/values.yaml
	openssl req -new -nodes -x509 -days 365 -keyout ./examples/certs/key.pem -out ./examples/certs/cert.pem -config ./sev-csc-demo/config/san.cnf
	sed -i 's#HOSTPATHPLACEHOLDER#$(shell pwd)/examples/certs/#g' ./sev-csc-demo/templates/vault.yaml
actix:
	docker build -t $(image_name) -f ./examples/actix.Dockerfile .
	docker push $(image_name)
	sed -i 's#IMAGENAMEPLACEHOLDER#$(image_name)#g' ./examples/actix/actix-knative.yaml
	sed -i 's/USERNAMEPLACEHOLDER/$(username)/g' ./examples/actix/actix-knative.yaml
	sed -i 's/PASSWORDPLACEHOLDER/$(password)/g' ./examples/actix/actix-knative.yaml
	vault write auth/userpass/users/$(username) password=$(password) policies=client
	kubectl apply -f ./examples/actix/actix-knative.yaml
flask:
	docker build -t $(image_name) -f ./examples/flask.Dockerfile .
	docker push $(image_name)
	sed -i 's#IMAGENAMEPLACEHOLDER#$(image_name)#g' ./examples/flask/flask-knative.yaml
	sed -i 's/USERNAMEPLACEHOLDER/$(username)/g' ./examples/flask/flask-knative.yaml
	sed -i 's/PASSWORDPLACEHOLDER/$(password)/g' ./examples/flask/flask-knative.yaml
	vault write auth/userpass/users/$(username) password=$(password) policies=client
	kubectl apply -f ./examples/flask/flask-knative.yaml
add-helm-repos:
	helm repo add projectcalico https://docs.tigera.io/calico/charts
	helm repo add metallb https://metallb.github.io/metallb
	helm repo add istio https://istio-release.storage.googleapis.com/charts
	helm repo add jetstack https://charts.jetstack.io
configure-vault:
	vault policy write client ./sev-csc-demo/config/policy.hcl
	vault auth enable userpass
	vault secrets enable kv-v2
	vault kv put -mount=kv-v2 secret-message secret-message="$(message)"
dependencies:
	helm install calico projectcalico/tigera-operator --version v3.25.1 --namespace tigera-operator --create-namespace --wait
	@if [ "$(load_balancer)" = "true" ]; then helm install metallb metallb/metallb  --version v0.13.10 -n metallb-system --create-namespace --wait && sed -i 's/IPPLACEHOLDER/$(ip)/g' ./sev-csc-demo/config/ip_pool.yaml && kubectl apply -f ./sev-csc-demo/config/ip_pool.yaml; fi
	helm install istio-base istio/base --version 1.18.0 -n istio-system --create-namespace
	helm install istiod istio/istiod --version 1.18.0  -n istio-system --set meshConfig.ingressService=istio-ingress --set meshConfig.ingressSelector=ingress --wait
	helm install istio-ingress istio/gateway --version 1.18.0  -n istio-system --create-namespace --wait
	@if [ "$(cert_manager)" = "true" ]; then helm install cert-manager jetstack/cert-manager --version v1.12.2 --namespace cert-manager --create-namespace --version v1.12.0 --set installCRDs=true;fi
.PHONY: actix clean add-helm-repos setup-cluster

