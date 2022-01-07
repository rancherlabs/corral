variable "corral_name" {}
variable "corral_user_id" {}
variable "corral_user_public_key" {}
variable "corral_public_key" {}

output "corral_node_pools" {
  value = {
    init = [
    for droplet in [digitalocean_droplet.controlplane[0]] : {
      name = droplet.name
      user = "root"
      address = droplet.ipv4_address
    }
    ]
  }
}
