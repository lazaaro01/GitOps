variable "container_name" {
  description = "Name of the container"
  type        = string
}

variable "image" {
  description = "Docker image to run"
  type        = string
}

variable "internal_port" {
  description = "Port exposed by the container"
  type        = number
  default     = 3000
}

variable "host_port" {
  description = "Port on the host to map"
  type        = number
  default     = 8080
}

variable "network_name" {
  description = "Docker network to attach to"
  type        = string
  default     = ""
}

variable "env_vars" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

resource "docker_image" "this" {
  name         = var.image
  keep_locally = true
}

resource "docker_container" "this" {
  name  = var.container_name
  image = docker_image.this.name

  dynamic "ports" {
    for_each = var.internal_port > 0 ? [1] : []
    content {
      internal = var.internal_port
      external = var.host_port
    }
  }

  dynamic "env" {
    for_each = var.env_vars
    content {
      name  = env.key
      value = env.value
    }
  }

  dynamic "networks_advanced" {
    for_each = var.network_name != "" ? [1] : []
    content {
      name = var.network_name
    }
  }
}

output "container_id" {
  value = docker_container.this.id
}

output "container_name" {
  value = docker_container.this.name
}

output "container_ip" {
  value = docker_container.this.network_data[0].ip_address
}
