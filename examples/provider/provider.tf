terraform {
  required_providers {
    aidbox = {
      source  = "reiinoldo/aidbox-license"
      version = "~> 0.1"
    }
  }
}

provider "aidbox" {
  # Set AIDBOX_API_TOKEN in the environment instead of hard-coding it here.
  endpoint = "https://aidbox.app/rpc"
}
