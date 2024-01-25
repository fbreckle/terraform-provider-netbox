terraform {
  required_providers {
    netbox = {
      source = "e-breuninger/netbox"
    }
  }
}

# example provider configuration for a local netbox deployment
# e.g. https://github.com/netbox-community/netbox-docker
#provider "netbox" {
#  server_url = "http://localhost:8000"
#  api_token  = "0123456789abcdef0123456789abcdef01234567"
#}

# example provider configuration for https://netboxdemo.om
provider "netbox" {
  server_url = "https://demo.netbox.dev"
  api_token  = "<your api token>"
}

