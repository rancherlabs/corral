variable "corral_name" {}
variable "corral_public_key" {}
variable "mapvar" {
  type = map(string)
}
variable "listvar" {
  type = list(number)
}

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

output "mapvar1" {
  value = var.mapvar
}

output "listvar1" {
  value = var.listvar
}