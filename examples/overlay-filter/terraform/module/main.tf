terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 2.13.0"
    }
  }
}

provider "docker" {}

resource "docker_volume" "data" {
  count = 2
  name = "${var.corral_name}-node-${count.index}"
}

resource "docker_container" "node" {
  count = 2
  image = "lscr.io/linuxserver/openssh-server"
  name  = "${var.corral_name}-node-${count.index}"

  ports {
    internal = 2222
  }

  env = [
    "PUBLIC_KEY=${var.corral_public_key}",
    "USER_NAME=corral",
    "USER_PASSWORD=corral",
    "SUDO_ACCESS=true",
    "PASSWORD_ACCESS=true",
  ]

  volumes {
    container_path = "/app"
    volume_name = docker_volume.data[count.index].name
  }
}