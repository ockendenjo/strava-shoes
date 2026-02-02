resource "aws_dynamodb_table" "bagging_db" {
  name                        = "strava-bagging-v2"
  billing_mode                = "PAY_PER_REQUEST"
  hash_key                    = "ID"
  table_class                 = "STANDARD"
  deletion_protection_enabled = false

  attribute {
    name = "ID"
    type = "S"
  }

  ttl {
    attribute_name = "Expiry"
    enabled        = true
  }
}
