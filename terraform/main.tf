module "lambda_function" {
  source = "terraform-aws-modules/lambda/aws"
  version = "~> 7.2"

  function_name = var.function_name
  description   = "Validate k8s external secrets by calling function url"
  handler       = "main"
  runtime       = "go1.x"

  create_package = false
  local_existing_package = data.null_data_source.downloaded_package.outputs["filename"]

  create_lambda_function_url = true

  attach_policy_jsons = true
  policy_jsons = [var.policy_json]
  number_of_policy_jsons = 1
}

locals {
  package_url = var.package_url
  downloaded  = "downloaded_package_${md5(local.package_url)}.zip"
}

resource "null_resource" "download_package" {
  triggers = {
    downloaded = local.downloaded
  }

  provisioner "local-exec" {
    command = "curl -L -o ${local.downloaded} ${local.package_url}"
  }
}

data "null_data_source" "downloaded_package" {
  inputs = {
    id       = null_resource.download_package.id
    filename = local.downloaded
  }
}
