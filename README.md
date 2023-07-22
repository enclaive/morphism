# morphisms
Kubernetes-based platform to build, deploy, and manage serverless confidential compute workloads. 
Documentation of the project can be found [here](https://morphisms.gitbook.io/morphisms-confidential-serverless-containers/) 
## AMD SEV-SNP
Under ./deployments/azure_dca you can find an Knative example using a Vault for storage of secrets. Here plain Knative is used, which does not support TLS for queue-proxy to serverless container communication.
Therefore, in the example, a message is stored in the vault, which can then be called up at an end point. To secure the communication you have the option to enable auto TLS or provide a wild-card certificate at 
the gateway. In-cluster traffic is secured by Istio mTLS, because internal encryption of Knative does not work currently.
## Intel SGX
Under ./deployments/azure_dcs you can find an Confidential Knative example using an enclaved Vault for storage of secrets. Here, an enclaved serverless container is deployed
which establishes a RA-TLS connection the Vault to receive a certificate and additional secrets. This certificate and the private key is then used by serverless container example for HTTPS.
Because, there is always a Proxy in front of the serverless container in Knative, you are not able to see this certificate. These proxies don't allow SSL Passthrough. See below, why
the Proxies do not have an TLS endpoint currently.

### Knative Fork for SGX

**What had to be adjusted?**

Queue-Proxy to serverless container traffic was only enabled for HTTP. This would allow to sniff the traffic to the enclave. 
Now, HTTPS is supported.

Knative uses a webhook, which checks all configuration of Knative Services send to the Kubernetes API. There, the source code had to be revised and a custom webhook is therefore used. This allows hostPaths and securityContext to be configured on a serverless container. Both are necessary to bring the SGX drivers into the serverless container. Currently the Intel Kubernetes Device Plugin does not work for DCAP on Azure and therefore this was necessary.

The Queue-Proxy is injected by Knative's Controller as Sidecar with a default configuration. This configuration had to be overridden so that the Kubernetes drivers are additionally mounted in the queue proxy. To operate an enclaved Queue Proxy, a custom Knative Controller is required. 
Currently, deploying an enclaved Queue-Proxy works, but additional configuration of Istio and Knative has to be done for this to work. There seems to be a timeout somewhere implemented. (This is Work-in-progress)

Enclaving the Webhook and Controller is not required, since they are only used for Deployment and the Kubernetes API is in the threat model included. Through remote attestation misconfiguration is detected.

Activator is enclaved and works. 

Currently, Gateway to Queue Proxy or Gateway to Activator does not support TLS in Knative. The Knative dev's are currently still working on this. The problem is that Knative dynamically changes routes to put the queue proxy or activator behind the gateway. Currently only HTTP routes are used here with the HTTP header modified by Istio. Istio does not offer this option for TLS. Tried to change the routes to TLS, which resulted in Knative not working anymore, because on the one hand a network controller, the proxies and the Autoscaler need the metrics. Therefore, the Activator as well as the queue proxy cannot currently have a TLS endpoint. Only in the route Activato to Queue-Proxy this would be possible, but this alone does not make sense. This will be implemented as soon as the Knative dev's are finished.
