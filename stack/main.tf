terraform {
  backend "s3" {
    # Backend configuration should be provided via backend-config or partial config
    # terraform init -backend-config="tfvars/backend-dev.hcl"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.0"
    }
  }
}

provider "aws" {
  allowed_account_ids = [var.aws_account_id]
  region              = var.aws_region

  default_tags {
    tags = {
      Environment = var.env
      Project     = "strava"
    }
  }
}
