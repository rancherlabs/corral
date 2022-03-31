variable "corral_name" {}
variable "corral_public_key" {}

output "corral_node_pools" {
  value = {
    all = [
    for n in docker_container.node : {
      name = n.name
      user = "corral"
      address = "127.0.0.1:${n.ports[0].external}"
    }
    ]
  }
}
