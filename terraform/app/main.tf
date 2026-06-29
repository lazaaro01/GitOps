terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
  }
}

provider "docker" {}

variable "app_name" {
  description = "Name of the application"
  type        = string
}

variable "image_tag" {
  description = "Docker image tag to deploy"
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

module "network" {
  source       = "../modules/network"
  network_name = "${var.app_name}-network"
}

module "container" {
  source         = "../modules/container"
  container_name = var.app_name
  image          = var.image_tag
  internal_port  = var.internal_port
  host_port      = var.host_port
  network_name   = module.network.network_name
}

output "app_name" {
  value = var.app_name
}

output "container_id" {
  value = module.container.container_id
}

output "container_name" {
  value = module.container.container_name
}

output "network_name" {
  value = module.network.network_name
}
