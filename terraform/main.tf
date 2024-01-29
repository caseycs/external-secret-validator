module "lambda_function" {
  source = "terraform-aws-modules/lambda/aws"
  version = "~> 7.2"

  function_name = var.function_name
  description   = "Validate k8s external secrets by calling function url"
  handler       = "main"
  runtime       = "go1.x"

  create_package = false
  local_existing_package = "../lambda-handler.zip"

  create_lambda_function_url = true

  attach_policy_jsons = true
  policy_jsons = [var.policy_json]
  number_of_policy_jsons = 1

  depends_on = [
    null_resource.this
  ]
}

resource "null_resource" "this" {
  provisioner "local-exec" {
    working_dir = path.module
    command = "GOOS=linux GOARCH=amd64 go build -o main ../; zip lambda-handler.zip main"
  }
}