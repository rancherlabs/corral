variable "corral_name" {} // name of the corral being created
variable "corral_user_id" {} // how the user is identified (usually github username)
variable "corral_user_public_key" {} // the users public key
variable "corral_public_key" {} // The corrals public key.  This should be installed on every node.
variable "corral_private_key" {} // The corrals private key.

// Package
variable "digitalocean_token" {}
variable "digitalocean_domain" {}
