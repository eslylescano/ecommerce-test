provider "aws" {
  region = "eu-west-1"
}

resource "aws_lambda_function" "example_lambda" {
  function_name = "ecommerce-test-ireland"
  role          = aws_iam_role.lambda_execution_role.arn
  handler       = "main.Handler"  # Replace with the actual handler name if different
  runtime       = "go1.x"
  filename      = "../../deployment.zip"  # Path to the ZIP file containing your Go code

  source_code_hash = filebase64sha256("../../deployment.zip")

  memory_size = 128
  timeout     = 10

  environment {
    variables = {
      key1 = "value1"
      key2 = "value2"
      # Add any environment variables your Go Lambda function requires
    }
  }
}

resource "aws_iam_role" "lambda_execution_role" {
  name = "lambda_execution_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com",
        },
      },
    ],
  })
}

resource "aws_iam_role_policy_attachment" "lambda_execution_policy_attachment" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda_execution_role.name
}

# Add any additional IAM policies that your Lambda function may need
