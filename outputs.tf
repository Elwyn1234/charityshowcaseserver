output "api_stage_base_url" {
  value = aws_apigatewayv2_stage.lambda.invoke_url
}

output "api_base_url" {
  value = aws_apigatewayv2_api.lambda.api_endpoint
}
