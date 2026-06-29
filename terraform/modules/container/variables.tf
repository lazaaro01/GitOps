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
