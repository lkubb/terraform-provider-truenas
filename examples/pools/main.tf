terraform {
  required_providers {
    truenas = {
      version = "0.1"
      source  = "dariusbakunas/providers/truenas"
    }
  }
}

provider "truenas" {}

data "truenas_pool_ids" "all" {}

output "all_pools" {
  value = data.truenas_pool_ids.all.ids
}
