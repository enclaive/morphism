path "kv-v2/data/secret-message"
{ 
capabilities = ["read"] 
}

path "pki_int/issue/flask-sgx" {

    capabilities = ["create", "update"]

}

path "pki/cert/ca" {

capabilities = ["read"]

}

path "auth/token/renew" {

    capabilities = ["update"]

}
    
path "auth/token/renew-self" {
    
    capabilities = ["update"]

}

