# SEV Example
##




## Requirements
- [Helm v3](https://helm.sh/docs/intro/install/)
- [Kubernetes (Kubelet, Kubectl, Kubeadm) ](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)
- a container runtime 
- Docker [Installation for Ubuntu](https://docs.docker.com/engine/install/ubuntu/) (for building and pushing the images) 
	- Add current user to docker group, if you haven't already: sudo usermod -aG docker $USER  and logout and login again
- Some Helm Charts take some time to install. Do not cancel them!
- [Vault-CLI](https://developer.hashicorp.com/vault/downloads)
- A confidential VM (protected by SEV or TDX). 
	- All steps have to be done inside a confidential VM
	- Security depends on the fact, that the control plane is inside a confidential VM
## Preparation
**1.1** Initialise a Kubernetes cluster and install a networking plugin of your choice. Here is an example using Calico (skip to [2.1](#load) if you want to install a different networking plugin, but do 1.2 first)

**1.2** 
Initialize your cluster. The pod-network-cidr is required by calico.

 ```bash
sudo kubeadm init --pod-network-cidr=192.168.0.0/16 
``` 

**1.3**
let the current user, use the kubeconfig:

```bash
  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
**1.4**
If you only have one node in your cluster, you need to use the control plane node to schedule pods on. Done by executing (!!not recommended in production!!):

 ```bash
kubectl taint nodes --all  node-role.kubernetes.io/control-plane-
``` 

## Set-up
This example uses the 'sed' command to replace some values. Meaning that, you have to pull the repository again, if you want to change any parameters you specified or manually change them.  
**2.1** The Makefile expects your shell to be in the same folder, otherwise some commands won't work. Add the required helm repos with:

 ```bash
make add-helm-repos
``` 
**2.2** 
Now for a quick-start you can issue the command below and replace the parameters to fit your likings (It is recommended to first read to 2.5 to understand, what the command does and then issue the quick-start command).
It issues the below make commands to 2.5 in a single command.
 ```bash
make quick-start ip=your_ip cert_manager=false load_balancer=true auto_tls=false domain=example.com development=true
``` 
Skip to [Vault](#vault_add), if you issued the command above.



Knative requires a load-balancer to work. The next step is different if you want to use an external load balancer, e.g. one from your cloud provider. In a cloud it is not required to use your cloud providers load-balancer. Should you want to use a domain, then this load-balancer must obtain the ip-address associated with your domain. For this example, it is not required to have a domain. Later, a workaround is shown. 

Should you want to use an external load balancer, then to not install an additional loadbalancer, set load_balancer=false in the next command. The ip parameter will then be ignored and does not have to be set. Refer to your cloud provider's documentation, when you are in a cloud. This must be pre-configured before issuing the next command, because the next command will wait and block until the load-balancer receives an ip.

Otherwise, you have to set load_balancer=true and specify an ip address, the load-balancer should obtain. This should be a public ip address, if you want to reach the vault or the example from an external network. Should this not be the case, then a private IP also works. In the cloud, the VM does not know the assigned public IP address. But, because your cloud provider translates the public ip to the private ip of your VM internally, you should use the private ip of your VM here and use the public ip to reach the load-balancer (domain should still be mapped to the public ip). 

Lastly, you can decide to install cert-manager into your cluster. Cert-manager is used for Knative's auto tls feature to automatically receive certificates from Let'sEncrypt for your newly deployed Knative Routes. For this to work, your load-balancer must be externally reachable. Use auto_tls=true to enable it and auto_tls=false to disable it. This make command can take several seconds until it's finished executing.

 ```bash
make dependencies ip=your-ip cert_manager=false load_balancer=true
``` 

<a name="demo"></a> 

**2.3** For https Vault requires a certificate and a private key. You can either use your own certificate and private key, if you want, or generate it with the next make command. With the latter, however, the following must be taken into account. The pre-main binary calls the vault with vault.vault.svc.cluster.local, and must either be changed in the main.go or included in the subject alternative name extension of your certificate. Moreover, also vault.{{your-domain.com}} should be included in the certificate. 

 ```bash
make cert domain=yourdomain.com
``` 
The certificates will be saved at ./examples/certs/ where the Vault and the examples expects them to be. The filenames should be cert.pem and key.pem. If you use your own certificate, the hostPath in the yaml of the vault
found at ./sev-csc-demo/templates/vault.yaml has to be set. For the vault to be able to read the private key, the permissions have to be adjusted.

**2.4** Additionally set domain=yourdomain.com in the make command below, if you did not have issued the command above. If you have set the auto_tls and the development variable to false, the DEMO expects a wildcard-certificate. This can be deployed to the cluster with the following command:

```bash
kubectl create --namespace istio-system secret tls wildcard-cert --key path/to/wildcard-key.pem --cert path/to/wildcard-cert.pem
``` 

Now you can install the Vault and Knative into your cluster by issuing.


  ```bash
make set-up auto_tls=false development=true
``` 

**2.5**
<a name="vault_add"></a> 
First, set the vault address environment variable for the vault-cli to know vault's address.
```bash
export VAULT_ADDR="https://vault.{{yourdomain.com}}"
``` 
Also, use to set the created certificate as trusted (this can also be done by copying the certificate to your OS's truststore):
```bash
 export VAULT_CACERT="./examples/certs/cert.pem"
``` 
**2.6** 
Initialize the vault and receive the unseal key + the admin token. See [DNS-Wildcard](#domain) if you do not have a domain, or add vault.{{yourdomain.com}} to your /etc/hosts file to map the ip address of the load balancer

```bash
vault operator init -key-shares=1 -key-threshold=1
``` 

Note that, it is important to keep both values confidential in production and have to be stored securely.
Unseal the vault by providing the unseal key.
```bash
vault operator unseal
``` 
Login to the vault by providing the token.

```bash
vault login
``` 

**2.7**
Almost done! Provide here a secret message, which should be displayed in the example. This secret gets saved in to the vault's key value store.
```bash
make configure-vault message={{your secret message}}
```

**2.8**
Lastly, deploy the example. For this to work, you have to push the image to a public image repository. A private one also works, but the Knative Service yaml files then have to include the imagePullSecret. Furthermore, this example assumes that the docker-cli is used, to push the image to the repository. To do this, you must be logged in to your docker account with the docker-cli. 

The next command will build and then push the image. Also, userpass is used as the vault authentication method and the user account is created here. Specify any username and password you want. 
For actix use:
```bash
make actix image_name=username/image_name username=user password=password
make flask image_name=username/image_name username=user password=password
```
If you have a domain, you have to create a wildcard-DNS entry for *.default.{{your-domain.com}} or an entry for actix.default.{{your-domain.com}} or flask.default.{{your-domain.com}}.
Then, you are able to see your secret message by calling:
```bash
 curl actix.default.{{your-domain.com}}/read_secret
 curl flask.default.{{your-domain.com}}/read-secret
```
### That's it, have fun with Serverless Confidential Containers
<a name="domain"></a>
If you do not have a domain, put your load-balancer ip and actix.default.{{your-domain.com}} in the /etc/hosts/ file. 
The section below describes how to create a wildcard DNS entry locally, if you wish to deploy multiple Knative services.
/etc/hosts/ does not support wildcard entries and you would have to manually add a new route everytime.
 


### This section is only required if you do not have a domain and wish to create a wildcard DNS entry locally. 
The following steps should be done on the **device that will be calling the web server**.

**3.1** 
Install dnsmasq. For Ubuntu use:
```bash
sudo apt-get install dnsmasq 
``` 

**3.2** 
Edit /etc/resolv.conf to use the dnsmasq DNS server, then add nameserver 127.0.0.1 to the first line.

**3.3**
Choose a domain and replace it with example.com here: address=/example.com/0.0.0.0
Replace 0.0.0.0 with the IP address you configured in the [Metallb Pool](#metallb).

**3.4** Restart dnsmasq
```bash
sudo systemctl restart dnsmasq
```

#### Knative always appends *.default as a prefix to the domain. If you enter example.com as the domain, Knative will create routes with the following format *.default.example.com, then you need to add a wildcard DNS entry for that exact format. For test.example.com this would be *.default.test.example.com. 



## What are the make commands doing?  (Set-up without make)
### make add-helm-repos and make dependencies
Kubernetes requires a CNI-plugin (Container Network Interface) for networking. Here, Calico is used, but this works with any CNI-plugin. Calico is an open-source CNI plug-in, widely used, easy to set-up and offers additional security features. Installation with Helm is done by issuing the following commands:

 ```bash
helm repo add projectcalico https://docs.tigera.io/calico/charts
helm install calico projectcalico/tigera-operator --version v3.25.1 --namespace tigera-operator --create-namespace --wait
``` 
<a name="load"></a>
Installs the metallb load-balancer:
 ```bash
helm repo add metallb https://metallb.github.io/metallb
helm install metallb metallb/metallb -n metallb-system --create-namespace --wait
``` 

<a name="metallb"></a>
Configures metallb to use obtain the specified ip. 

Now use the following command, but replace {{ Enter your ip here }} with your IP.
 ```bash
kubectl apply -f - <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: first-pool
  namespace: metallb-system
  annotations:
    "helm.sh/hook": "post-install"
    "helm.sh/hook-weight": "0"
spec:
  addresses:
  - {{ Enter your ip here}} - {{ Enter your ip here }}
EOF
``` 
<a name="set_up"></a> 

For HTTP-01 challenge verification, Cert-Manager uses an Istio ingress and for this to work, Istiod must be configured with meshConfig.ingressService=istio-ingress --set meshConfig.ingressSelector=ingress to actually use the ingress for incoming traffic.  Knative relies on a network layer to provide load balancing. This demo uses Istio to do this. 

```bash
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm install istio-base istio/base -n istio-system --create-namespace
helm install istiod istio/istiod -n istio-system --set meshConfig.ingressService=istio-ingress --set meshConfig.ingressSelector=ingress --wait
helm install istio-ingress istio/gateway -n istio-system --create-namespace --wait
``` 

Installs cert-manager
```bash
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.12.0 --set installCRDs=true
``` 
<a name="demo"></a> 
## Deploy the Demo

### make certs
make certs generates with openssl a self-signed certificate for the vault to use. This certificate is also distributed into the docker image of the examples.
### make set-up
Sets the passed parameters into the values.yaml, packages and installs helm repository. This installs pre-configured Knative and the Vault
### make configure-vault
Writes a policy, which allows to read the secret-message secret, when binded to a userapass. Enables userpass and the key-value store. Additionally, the message is put into the key-value store.
### make actix or make flask
Builds and pushes the image and writes the image name, the username and the password into the yaml file for Kubernetes. Also, the username and the password is created as user-account. Lastly, the yaml file is applied.
