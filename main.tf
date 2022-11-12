# Set envrionment variables from .env file
locals {
  env = { for tuple in regexall("(.*)=(.*)", file(".env")) : tuple[0] => tuple[1] }
}

/* 
  * * Provider configuration
 */
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}


/* 
  * * Data sources
 */
data "archive_file" "rental_bot_messenger_webhook" {
  type        = "zip"
  source_file = "rental_bot_messenger_webhook"
  output_path = "rental_bot_messenger_webhook.zip"
}


/*
  * * Lambda function
*/
resource "aws_lambda_function" "rental_bot_messenger_webhook" {
  filename      = "rental_bot_messenger_webhook.zip"
  function_name = "rental_bot_messenger_webhook"
  handler       = "main"
  role          = aws_iam_role.iam_for_rental_bot_messenger_webhook_lambda.arn

  source_code_hash = data.archive_file.rental_bot_messenger_webhook.output_base64sha256

  runtime = "go1.x"

  environment {
    variables = {
      VERIFY_TOKEN      = local.env["VERIFY_TOKEN"]
      PAGE_ACCESS_TOKEN = local.env["PAGE_ACCESS_TOKEN"]
      PAGE_ID           = local.env["PAGE_ID"],
      SNS_TOPIC_ARN     = local.env["SNS_TOPIC_ARN"]
    }
  }
}

resource "aws_lambda_function_url" "rental_bot_messenger_webhook" {
  function_name      = aws_lambda_function.rental_bot_messenger_webhook.function_name
  authorization_type = "NONE"
}

/*
  * * SNS topic
*/
resource "aws_sns_topic" "rental_bot_messenger_webhook" {
  name = "rental_bot_messenger_webhook"
}

resource "aws_sns_topic_subscription" "rental_bot_messenger_webhook" {
  topic_arn = aws_sns_topic.rental_bot_messenger_webhook.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.rental_bot_messenger_webhook.arn
}

/*
  * * IAM role and policies
*/


resource "aws_iam_role" "iam_for_rental_bot_messenger_webhook_lambda" {
  name = "iam_for_rental_bot_messenger_webhook_lambda"

  # Generates temporary security credentials for the IAM role to use
  assume_role_policy = jsonencode(
    {
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Action" : "sts:AssumeRole",
          "Principal" : {
            "Service" : "lambda.amazonaws.com"
          },
          "Effect" : "Allow",
          "Sid" : ""
        }
      ]
  })

  # AWSLambdaBasicExecutionRole for CloudWatch Logs
  managed_policy_arns = ["arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole", aws_iam_policy.lambda_publish_to_sns.arn]

}
resource "aws_iam_policy" "lambda_publish_to_sns" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : [
          "sns:Publish"
        ],
        "Resource" : [
          local.env["SNS_TOPIC_ARN"]
        ],
        "Effect" : "Allow"
      }
    ]
  })
}


