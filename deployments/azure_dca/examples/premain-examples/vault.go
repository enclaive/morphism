// File from: https://github.com/hashicorp/vault-examples/blob/main/examples/token-renewal/go/example.go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/userpass"
)

func login(client *vault.Client) (*vault.Secret, error) {
	// WARNING: A plaintext password like this is obviously insecure.
	// See the files in the auth-methods directory for full examples of how to securely
	// log in to Vault using various auth methods. This function is just
	// demonstrating the basic idea that a *vault.Secret is returned by
	// the login call.
	username := os.Getenv("user")
	password := os.Getenv("password")
	userpassAuth, err := auth.NewUserpassAuth(username, &auth.Password{FromString: password})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize userpass auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), userpassAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to userpass auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}
