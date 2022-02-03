variable "agent_count" {
  type = number
}
variable "agent_size" {
}

resource "digitalocean_droplet" "agent" {
  count = var.agent_count

  name = "${var.corral_user_id}-${random_id.cluster_id.hex}-agent-${count.index}"
  image    = "ubuntu-20-04-x64"
  region   = "sfo3"
  size     = var.agent_size
  tags = [var.corral_user_id, var.corral_name]
  ssh_keys = [digitalocean_ssh_key.corral_key.id]
}