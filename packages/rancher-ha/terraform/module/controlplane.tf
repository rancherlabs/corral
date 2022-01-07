resource "digitalocean_droplet" "node" {
  count = 3

  name = "${var.corral_user_id}-${random_id.cluster_id.hex}-cp-${count.index}"
  image    = "ubuntu-20-04-x64"
  region   = "sfo3"
  size     = "s-2vcpu-4gb"
  tags = [var.corral_user_id, var.corral_name]
  ssh_keys = [digitalocean_ssh_key.corral_key.id]
}

resource "random_id" "rancher_host" {
  byte_length       = 6
}

resource "digitalocean_record" "rancher_host" {
  domain = var.digitalocean_domain
  name   = random_id.rancher_host.hex
  type   = "A"
  value  = digitalocean_droplet.node[0].ipv4_address
}

resource "digitalocean_record" "kube_api" {
  domain = var.digitalocean_domain
  name   = random_id.cluster_id.hex
  type   = "A"
  value  = digitalocean_droplet.node[0].ipv4_address
}

output "kube_api_host" {
  value = join(".", [digitalocean_record.kube_api.name, digitalocean_record.kube_api.domain])
}

output "rancher_host" {
  value = join(".", [digitalocean_record.rancher_host.name, digitalocean_record.rancher_host.domain])
}
