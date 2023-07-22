SECRET='{"environment": {}, "files": {"/dev/attestation/keys/data":"c1ydwRokay1R4xZ3mPwd1w==","/dev/attestation/keys/logs":"nKz4dRYLWQBhkW9bzs6HQw=="}, "argv": []}'
quick-start:
	make dependencies ip=$(ip) load_balancer=$(load_balancer)
	make set-up development=$(development) domain=$(domain) pccs=$(pccs) api_key=$(api_key)
set-up:
	sed -i 's/DOMAINPLACEHOLDER/$(domain)/g' ./scsc-demo/values.yaml
	sed -i 's/DEVELOPMENTPLACEHOLDER/$(development)/g' ./scsc-demo/values.yaml
	sed -i 's/PCCSPLACEHOLDER/$(pccs)/g' ./scsc-demo/values.yaml
	@if [ "$(pccs)" = "true" ]; then sed -i 's#APIKEYPLACEHOLDER#$(api_key)#g' ./scsc-demo/values.yaml; fi
	helm package scsc-demo
	helm install scsc-demo scsc-demo-0.1.0.tgz
actix:
	docker build -t $(image_name) -f ./examples/actix.Dockerfile --build-arg projectName=test_serverless .
	docker push $(image_name)
	sed -i 's#IMAGENAMEPLACEHOLDER#$(image_name)#g' ./examples/actix-sgx/actix-knative.yaml
	make add-enclave name=actix-sgx domain=$(domain) measurement=40aae473bbc5963693bd3a230d70dd38794daef04edd9a8ec50ba58078ecc457
	kubectl apply -f ./examples/actix-sgx/actix-knative.yaml
flask:
	docker build -t $(image_name) -f ./examples/flask.Dockerfile .
	docker push $(image_name)
	sed -i 's#IMAGENAMEPLACEHOLDER#$(image_name)#g' ./examples/flask-sgx/flask-knative.yaml
	make add-enclave name=flask-sgx domain=$(domain) measurement=68f02e96a307ba2d9544ba6569bc8b8c81789c0455227f3f659f2e86df85810a
	kubectl apply -f ./examples/flask-sgx/flask-knative.yaml
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
	vault secrets enable -path=sgx-pki pki  
	vault secrets tune --max-lease-ttl=87600h sgx-pki
	vault write -format=json sgx-pki/intermediate/generate/internal common_name="$(domain) Intermediate Authority" | jq -r '.data.csr' > pki_intermediate.csr 
	sed -i 's/DOMAINPLACEHOLDER/$(domain)/g' ./examples/actix-sgx/actix-knative.yaml
	mkdir -p ./examples/certs
	vault write -format=json pki/root/sign-intermediate csr=@pki_intermediate.csr format=pem_bundle ttl="43800h" | jq -r '.data.certificate' > ./examples/certs/intermediate.cert.pem 
	vault write sgx-pki/intermediate/set-signed certificate=@./examples/certs/intermediate.cert.pem
	vault write sgx-pki/roles/vault-$(domain) allowed_domains="$(domain)" allow_subdomains=true max_ttl="720h" 
	vault secrets enable -path=sgx-app kv-v2 
add-enclave:
	bash -c "vault policy write sgx-app/"$(name)" - < <(env -i NAME="$(name)" envsubst < ./scsc-demo/config/vault.sgx.policy.template)"
	vault kv put -format=json -mount=sgx-app "$(name)" provision=$(SECRET)
	vault write -format=json auth/sgx-auth/enclave/$(name) mrenclave="$(measurement)"
	vault write -format=json sgx-pki/roles/$(name) allowed_domains="default.$(domain)" allow_subdomains=true allow_localhost=true ttl=8760h key_type="ec" key_bits="256"
dependencies:
	helm install calico projectcalico/tigera-operator --version v3.25.1 --namespace tigera-operator --create-namespace --wait
	@if [ "$(load_balancer)" = "true" ]; then helm install metallb metallb/metallb  --version v0.13.10 -n metallb-system --create-namespace --wait && sed -i 's/IPPLACEHOLDER/$(ip)/g' ./scsc-demo/config/ip_pool.yaml && kubectl apply -f ./scsc-demo/config/ip_pool.yaml; fi
	helm install istio-base istio/base --version 1.18.0 -n istio-system --create-namespace
	helm install istiod istio/istiod --version 1.18.0  -n istio-system --set meshConfig.ingressService=istio-ingress --set meshConfig.ingressSelector=ingress --wait
	helm install istio-ingress istio/gateway --version 1.18.0  -n istio-system --create-namespace --wait
.PHONY: actix clean add-helm-repos setup-cluster

