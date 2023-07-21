current_dir = $(shell pwd)
quick-start:
	make dependencies ip=$(ip) load_balancer=$(load_balancer)
	make set-up auto_tls=$(auto_tls) development=$(development)
set-up:
	sed -i 's/AUTOTLSPLACEHOLDER/$(auto_tls)/g' ./sev-csc-demo/values.yaml
	@if [ -n "$(domain)" ]; then sed -i 's/DOMAINPLACEHOLDER/$(domain)/g' ./sev-csc-demo/values.yaml; fi
	sed -i 's/DEVELOPMENTPLACEHOLDER/$(development)/g' ./sev-csc-demo/values.yaml
	helm package sev-csc-demo
	helm install sev-csc-demo sev-csc-demo-0.1.0.tgz
actix:
	docker build -t $(image_name) -f ./examples/actix.Dockerfile .
	docker push $(image_name)
	sed -i 's#IMAGENAMEPLACEHOLDER#$(image_name)#g' ./examples/actix/actix-knative.yaml
	sed -i 's/USERNAMEPLACEHOLDER/$(username)/g' ./examples/actix/actix-knative.yaml
	sed -i 's/PASSWORDPLACEHOLDER/$(password)/g' ./examples/actix/actix-knative.yaml
	make add-enclave name=actix-sgx domain=$(domain) measurement=f84f225e8b599517dbc4a75bae84cd598c79ee890ee10dff5a538689d5306a2d
	kubectl apply -f ./examples/actix/actix-knative.yaml
flask:
	docker build -t $(image_name) -f ./examples/flask.Dockerfile .
	docker push $(image_name)
	sed -i 's#IMAGENAMEPLACEHOLDER#$(image_name)#g' ./examples/flask/flask-knative.yaml
	sed -i 's/USERNAMEPLACEHOLDER/$(username)/g' ./examples/flask/flask-knative.yaml
	sed -i 's/PASSWORDPLACEHOLDER/$(password)/g' ./examples/flask/flask-knative.yaml
	make add-enclave name=flask-sgx domain=$(domain) measurement=f84f225e8b599517dbc4a75bae84cd598c79ee890ee10dff5a538689d5306a2d
	kubectl apply -f ./examples/flask/flask-knative.yaml
add-helm-repos:
	helm repo add projectcalico https://docs.tigera.io/calico/charts
	helm repo add metallb https://metallb.github.io/metallb
	helm repo add istio https://istio-release.storage.googleapis.com/charts
	helm repo add jetstack https://charts.jetstack.io
configure-vault:
	vault plugin register -sha256="f4a2ad37c5177baaaf8559a80a2edca3e158c1b9161e1274e7289e6628d1745e" auth vault-plugin-auth-sgx
	vault auth enable -path=sgx-auth vault-plugin-auth-sgx
	vault secrets enable pki
	vault secrets tune --max-lease-ttl=87600h pki
	vault write -field=certificate pki/root/generate/internal issuing_certificates="https://vault.$(domain)/v1/pki/crl" crl_distribution=_points="https://vault.$(domain)/v1/pki/crl"
	vault secrets enable -path=pki_int pki  
	vault secrets tune --max-lease-ttl=87600h pki_int
	vault write -format=json pki_int/intermediate/generate/internal common_name="$(domain) Intermediate Authority" | jq -r '.data.csr' > pki_intermediate.csr 
	mkdir -p ./examples/certs
	vault write -format=json pki/root/sign-intermediate csr=@pki_intermediate.csr format=pem_bundle ttl="43800h" | jq -r '.data.certificate' > ./examples/certs/intermediate.cert.pem
	vault write pki_int/intermediate/set-signed certificate=@intermediate.cert.pem
	vault write pki_int/roles/vault-$(domain) allowed_domains="$(domain)" allow_subdomains=true max_ttl="720h" 
  	vault secrets enable -path=sgx-app kv-v2 
add-enclave:
	vault kv put -format=json -mount=sgx-app "${name}" secret-message="${message}"
	vault policy write sgx-app/"${name}" ./scsc-demo/config/$(name)-policy.hcl
	vault write -format=json auth/sgx-auth/enclave/$(name) mrenclave="$(measurement)" policies=$(name)-policy
	vault write -format=json pki_int/roles/actix-sgx allowed_domains="$(name).default.$(domain)" allow_subdomains=true allow_localhost=true ttl=8760h key_type="ec" key_bits="256"
dependencies:
	helm install calico projectcalico/tigera-operator --version v3.25.1 --namespace tigera-operator --create-namespace --wait
	@if [ "$(load_balancer)" = "true" ]; then helm install metallb metallb/metallb  --version v0.13.10 -n metallb-system --create-namespace --wait && sed -i 's/IPPLACEHOLDER/$(ip)/g' ./sev-csc-demo/config/ip_pool.yaml && kubectl apply -f ./sev-csc-demo/config/ip_pool.yaml; fi
	helm install istio-base istio/base --version 1.18.0 -n istio-system --create-namespace
	helm install istiod istio/istiod --version 1.18.0  -n istio-system --set meshConfig.ingressService=istio-ingress --set meshConfig.ingressSelector=ingress --wait
	helm install istio-ingress istio/gateway --version 1.18.0  -n istio-system --create-namespace --wait
.PHONY: actix clean add-helm-repos setup-cluster

