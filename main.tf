#####################################################################################
# IAM Role for Lambda
#####################################################################################

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "iam_for_lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}






#####################################################################################
# Lambda Function - hello-world2
#####################################################################################

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "build/lambdas/lambda1"
  output_path = "build/lambdas/lambda1.zip"
}

resource "aws_lambda_function" "hello-world2" {
  # If the file is not in the current working directory you will need to include a
  # path.module in the filename.
  filename      = "build/lambdas/lambda1.zip"
  source_code_hash = data.archive_file.lambda.output_base64sha256
  function_name = "hello-world2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "lambda1"
  runtime = "go1.x"
}

resource "aws_lambda_function_url" "test_latest" {
  function_name      = aws_lambda_function.hello-world2.function_name
  authorization_type = "NONE"
}






#####################################################################################
# API Gateway
#####################################################################################

resource "aws_apigatewayv2_api" "lambda" {
  name          = "charity_showcase_api_gateway"
  protocol_type = "HTTP"
}

resource "aws_cloudwatch_log_group" "hello-world2" {
  name = "/aws/lambda/hello-worldddddddd2"

  retention_in_days = 30
}

resource "aws_apigatewayv2_stage" "lambda" {
  api_id = aws_apigatewayv2_api.lambda.id
  name        = "hello-world2"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.hello-world2.arn
    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
      }
    )
  }
}

resource "aws_apigatewayv2_integration" "hello-world2" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = aws_lambda_function.hello-world2.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "hello-world2" {
  api_id = aws_apigatewayv2_api.lambda.id

  route_key = "POST /hello"
  target    = "integrations/${aws_apigatewayv2_integration.hello-world2.id}"
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.hello-world2.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}

