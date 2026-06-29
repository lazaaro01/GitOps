variable "volume_name" {
  description = "Name of the Docker volume"
  type        = string
}

resource "docker_volume" "this" {
  name = var.volume_name
}

output "volume_id" {
  value = docker_volume.this.id
}

output "volume_name" {
  value = docker_volume.this.name
}

output "mount_point" {
  value = docker_volume.this.mountpoint
}
