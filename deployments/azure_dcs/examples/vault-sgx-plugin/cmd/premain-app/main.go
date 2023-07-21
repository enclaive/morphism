package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"app/premain/attest"

	vault "github.com/hashicorp/vault/api"
	"golang.org/x/sys/unix"
)

const (
	EnvEnclaveName   = "ENCLAIVE_NAME"
	EnvEnclaveServer = "ENCLAIVE_SERVER"
	mountPath        = "auth/sgx-auth/login"
	// VaultMount default mount path for KV v2 in dev mode
	VaultMount = "sgx-app"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type SgxAuth struct {
	request *attest.Request
}

func envConfig(name string) string {
	value, ok := os.LookupEnv(name)

	if !ok {
		panic(fmt.Errorf("environment variable '%s' missing", name))
	}

	return value
}

func vaultClient(certificate *tls.Certificate) *vault.Client {
	config := vault.DefaultConfig()

	config.Address = envConfig(EnvEnclaveServer)

	var peerCertificates [][]byte = nil
	transport := config.HttpClient.Transport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = true
	transport.TLSClientConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if peerCertificates == nil {
			peerCertificates = rawCerts
		} else {
			for i, rawCert := range rawCerts {
				if !bytes.Equal(peerCertificates[i], rawCert) {
					return fmt.Errorf("peer certificate '%d' changed", i)
				}
			}
		}

		return nil
	}
	transport.TLSClientConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return certificate, nil
	}

	config.HttpClient.Transport = transport

	client, err := vault.NewClient(config)
	check(err)

	return client
}
func NewSgxAuth(request *attest.Request) *SgxAuth {
	return &SgxAuth{request: request}
}
func (s *SgxAuth) Login(ctx context.Context, client *vault.Client) (*vault.Secret, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	attestation, err := json.Marshal(s.request)
	if err != nil {
		return nil, err
	}

	loginData := map[string]interface{}{
		"id":          s.request.Name,
		"attestation": attestation,
	}

	resp, err := client.Logical().WriteWithContext(ctx, mountPath, loginData)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func vaultRequest(client *vault.Client, request *attest.Request) (*attest.Secrets, *attest.TlsConfig) {

	_, err := client.Auth().Login(context.Background(), NewSgxAuth(request))
	check(err)

	kvSecret, err := client.KVv2(VaultMount).Get(context.Background(), request.Name)
	check(err)

	secrets := new(attest.Secrets)
	check(json.Unmarshal([]byte(kvSecret.Data["provision"].(string)), secrets))

	path := fmt.Sprintf("sgx-pki/issue/%s", request.Name)
	domain := os.Getenv("domain")
	tlsSecret, err := client.Logical().WriteWithContext(context.Background(), path, map[string]interface{}{
		"common_name": fmt.Sprintf("%s.default.%s", request.Name, domain),
		"format":      "der",
	})
	check(err)

	tlsConfig := new(attest.TlsConfig)
	tlsRaw, err := json.Marshal(tlsSecret.Data)
	check(err)

	check(json.Unmarshal(tlsRaw, tlsConfig))

	return secrets, tlsConfig
}

func secretsProvision(secrets *attest.Secrets) {
	args := make([]string, len(os.Args))
	copy(args, os.Args)
	args = append(args, secrets.Argv...)
	args[0] = filepath.Base(args[0])

	for k, v := range secrets.Environment {
		check(os.Setenv(k, v))
	}

	//FIXME gramine encrypted mount keys must be written first
	//for path, content := range secrets.Files {
	//	check(os.WriteFile(path, content, 0600))
	//}

	check(unix.Exec(os.Args[0], args, os.Environ()))
}

func main() {
	enclaveName := envConfig(EnvEnclaveName)

	privateKey, err := attest.GenerateEcKey()
	check(err)

	tlsCtx := attest.NewTlsContext(privateKey, enclaveName)

	selfSignedCertificate, err := attest.GenerateCert(tlsCtx)
	check(err)

	rawQuote, err := attest.NewGramineIssuer().Issue(selfSignedCertificate.Raw)
	check(err)

	client := vaultClient(&tls.Certificate{
		Certificate: [][]byte{selfSignedCertificate.Raw},
		PrivateKey:  privateKey,
	})

	secrets, tlsConfig := vaultRequest(client, &attest.Request{
		Name:  enclaveName,
		Quote: rawQuote,
	})

	check(tlsConfig.Save("/secrets/tmp"))

	secretsProvision(secrets)
}
