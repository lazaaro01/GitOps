variable "network_name" {
  description = "Name of the Docker network"
  type        = string
}

resource "docker_network" "this" {
  name = var.network_name
  driver = "bridge"
}

output "network_id" {
  value = docker_network.this.id
}

output "network_name" {
  value = docker_network.this.name
}
