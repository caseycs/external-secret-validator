module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "~> 7.2"

  function_name = var.function_name
  description   = "Validate k8s external secrets by calling function url"
  handler       = "main"
  runtime       = "go1.x"

  create_package         = false
  local_existing_package = "${path.cwd}/handler_${trimspace(data.local_file.md5.content)}.zip"

  create_lambda_function_url = true

  attach_policy_jsons    = true
  policy_jsons           = [var.policy_json]
  number_of_policy_jsons = 1
}

resource "null_resource" "build_lambda_handler" {
  triggers = {
    go_sum = filemd5("${path.module}/../go.sum")
    main_go = filemd5("${path.module}/../main.go")
    verify_go = filemd5("${path.module}/../verify.go")
  }

  provisioner "local-exec" {
    working_dir = "${abspath(path.module)}/../"
    command = <<EOT
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .;
      zip ${path.cwd}/handler.zip main;
      MD5=$(md5 -q ${path.cwd}/handler.zip | tee ${path.cwd}/handler.zip.md5);
      cp ${path.cwd}/handler.zip ${path.cwd}/handler_$MD5.zip
    EOT
  }
}

data "local_file" "md5" {
  filename = "${path.cwd}/handler.zip.md5"
  depends_on = [null_resource.build_lambda_handler]
}
