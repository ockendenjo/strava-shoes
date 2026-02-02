resource "aws_ssm_parameter" "client_id" {
  name  = "/strava/clientId"
  type  = "String"
  value = "placeholder"
  tier  = "Standard"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "client_secret" {
  name  = "/strava/clientSecret"
  type  = "String"
  value = "placeholder"
  tier  = "Standard"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "access_token" {
  name  = "/strava/accessToken"
  type  = "String"
  value = "placeholder"
  tier  = "Standard"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "refresh_token" {
  name  = "/strava/refreshToken"
  type  = "String"
  value = "placeholder"
  tier  = "Standard"

  lifecycle {
    ignore_changes = [value]
  }
}
