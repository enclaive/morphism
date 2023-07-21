path "pki_int/issue/knative" {
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

