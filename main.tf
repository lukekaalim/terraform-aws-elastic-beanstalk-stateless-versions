resource "example_server" "my-server2" {
  application_name = "1.2.3.4"
  application_store_bucket_name = "lk-sandbox"
  application_version_filename = "local.txt"
}

output "application_version" {
  value = "${example_server.my-server2.application_name}"
}

provider "example" {
  region = "ap-southeast-2"
}