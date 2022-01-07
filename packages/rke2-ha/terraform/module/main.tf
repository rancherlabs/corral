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

variable "digitalocean_token" {}
variable "digitalocean_domain" {}

provider "digitalocean" {
  token = var.digitalocean_token
}

resource "random_id" "cluster_id" {
  byte_length       = 6
}

resource "digitalocean_ssh_key" "corral_key" {
  name       = "corral-{var.corral_user_id}-${random_id.cluster_id.hex}"
  public_key = var.corral_public_key
}
