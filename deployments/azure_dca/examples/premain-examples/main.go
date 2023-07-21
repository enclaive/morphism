package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	vault "github.com/hashicorp/vault/api"
)

func main() {
	log.Default().Print("started pre-main")
	config := vault.DefaultConfig()
	config.Address = "https://vault.vault.svc.cluster.local:8200"
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	_, err = login(client)
	if err != nil {
		log.Fatalf("Error while logging in %v", err)
	}

	secret, err := client.KVv2("kv-v2").Get(context.Background(), "secret-message")
	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}
	value, ok := secret.Data["secret-message"].(string)
	fmt.Printf("secret: %v retrieved \n", value)

	if !ok {
		log.Fatalf("wrong secret format %v", err)
	}
	binary_path := os.Getenv("path")
	args := os.Args
	args = args[1:]
	if binary_path == "python3" {
		os.Setenv("secret-message", value)
	} else {
		args = append(args, value)
	}

	output, err := exec.Command(binary_path, args...).Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(output))
}
