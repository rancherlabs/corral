variable "controlplane_count" {
  type = number
  default = 1
}
variable "controlplane_size" {
  default = "s-2vcpu-4gb"
}

resource "digitalocean_droplet" "controlplane" {
  count = var.controlplane_count

  name = "${var.corral_user_id}-${random_id.cluster_id.hex}-cp-${count.index}"
  image    = "ubuntu-20-04-x64"
  region   = "sfo3"
  size     = var.controlplane_size
  tags = [var.corral_user_id, var.corral_name]
  ssh_keys = [digitalocean_ssh_key.corral_key.id]
}

resource "digitalocean_record" "kube_api" {
  domain = var.digitalocean_domain
  name   = random_id.cluster_id.hex
  type   = "A"
  value  = digitalocean_droplet.controlplane[0].ipv4_address
}

output "kube_api_host" {
  value = join(".", [digitalocean_record.kube_api.name, digitalocean_record.kube_api.domain])
}