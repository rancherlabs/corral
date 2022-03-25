variable "corral_name" {}
variable "corral_user_id" {}
variable "corral_user_public_key" {}
variable "corral_public_key" {}

output "corral_node_pools" {
  value = {
    a = [
    for n in [docker_container.node[0]] : {
      name = n.name
      user = "corral"
      address = "127.0.0.1:${n.ports[0].external}"
    }
    ]
    b = [
    for n in [docker_container.node[1]] : {
      name = n.name
      user = "corral"
      address = "127.0.0.1:${n.ports[0].external}"
    }
    ]
  }
}
