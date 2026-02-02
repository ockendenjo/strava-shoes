data "aws_s3_bucket" "lambda_binaries" {
  bucket = var.lambda_binaries_bucket
}

data "aws_s3_object" "manifest_file" {
  bucket = data.aws_s3_bucket.lambda_binaries.id
  key    = "lambda_manifests/default.json"
}

locals {
  manifest = jsondecode(data.aws_s3_object.manifest_file.body)
}
