variable "aws_account_id" {
  description = "AWS Account ID"
  type        = string
}

variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "eu-west-1"
}

variable "client_id" {
  type        = number
  description = "Strava app client ID"
  default     = 0
}

variable "env" {
  description = "Environment name (dev or pro)"
  type        = string

  validation {
    condition     = contains(["dev", "pro"], var.env)
    error_message = "Environment must be either 'dev' or 'pro'."
  }
}

variable "gear_ids" {
  type        = string
  description = "Stringified JSON of gear IDs to warn about"
  default     = "[\"g9558316\", \"\"]"
}

variable "lambda_binaries_bucket" {
  type = string
}

variable "permissions_boundary_arn" {
  description = "ARN of the IAM permissions boundary policy"
  type        = string
}
