terraform {
  required_version = ">= 0.13"
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "random" {}
provider "digitalocean" {
  token = var.digitalocean_token
}

// it is best practice to distinguish an environment with a random id to avoid collisions
resource "random_id" "registry_id" {
  byte_length       = 6
}

// we will use the corral public key to get access to nodes to provision them later
resource "digitalocean_ssh_key" "corral_key" {
  name       = "${var.corral_user_id}-${random_id.registry_id.hex}"
  public_key = var.corral_public_key
}

resource "digitalocean_droplet" "registry" {
  count = 1

  name = "${var.corral_user_id}-${random_id.registry_id.hex}-registry"
  image    = "ubuntu-20-04-x64"
  region   = "sfo3"
  size     = "s-1vcpu-2gb"
  tags = [var.corral_user_id, random_id.registry_id.hex] // when possible resources should be marked with the associated corral
  ssh_keys = [digitalocean_ssh_key.corral_key.id]
}

resource "digitalocean_record" "registry" {
  domain = var.digitalocean_domain
  name   = random_id.registry_id.hex
  type   = "A"
  value  = digitalocean_droplet.registry[0].ipv4_address
}