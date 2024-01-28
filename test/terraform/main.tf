resource "aws_secretsmanager_secret" "this" {
  name = "external_secret_validator_test"
}

resource "aws_secretsmanager_secret_version" "this" {
  secret_id     = aws_secretsmanager_secret.this.id
  secret_string = jsonencode({
    int_key = 1234
    string_key = "string_key_value"
  })
}