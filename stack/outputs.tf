output "auth_callback_domain" {
  description = "Auth callback domain"
  value       = replace(aws_apigatewayv2_api.http_api.api_endpoint, "https://", "")
}

output "strava_auth_url" {
  description = "Strava authentication URL"
  value       = "https://www.strava.com/oauth/authorize?client_id=${var.client_id}&response_type=code&scope=activity:read&redirect_uri=${aws_apigatewayv2_stage.default.invoke_url}/auth"
}
