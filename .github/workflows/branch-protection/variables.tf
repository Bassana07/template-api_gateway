variable "github_token" {
  description = "GitHub Personal Access Token"
  sensitive   = true
}

variable "repository_name" {
  description = "stage_tokken"
  type        = string
}

variable "repository_owner" {
  description = "cie-site-web"
  type        = string
}