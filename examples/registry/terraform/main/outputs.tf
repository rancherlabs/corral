output "corral_node_pools" {
  value = {
    registry = [
    for droplet in digitalocean_droplet.registry : {
      name = droplet.name // unique name of node
      user = "root" // ssh username
      address = droplet.ipv4_address // address of ssh host
    }
    ]
  }
}

output "registry_host" {
  value = join(".", [digitalocean_record.registry.name, digitalocean_record.registry.domain])
}