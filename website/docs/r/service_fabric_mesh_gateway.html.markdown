---
subcategory: "Service Fabric Mesh"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_service_fabric_mesh_gateway"
description: |-
  Manages a Service Fabric Mesh Gateway.
---

# azurerm_service_fabric_mesh_gateway

Manages a Service Fabric Mesh Gateway.

## Example Usage


```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_service_fabric_mesh_gateway" "example" {
  name                = "example-mesh-gateway"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Service Fabric Mesh Gateway. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the Resource Group in which the Service Fabric Mesh Gateway exists. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the Azure Region where the Service Fabric Mesh Gateway should exist. Changing this forces a new resource to be created.

* `description` - (Optional) A description of this Service Fabric Mesh Gateway.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Service Fabric Mesh Gateway.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Service Fabric Mesh Gateway.
* `update` - (Defaults to 30 minutes) Used when updating the Service Fabric Mesh Gateway.
* `read` - (Defaults to 5 minutes) Used when retrieving the Service Fabric Mesh Gateway.
* `delete` - (Defaults to 30 minutes) Used when deleting the Service Fabric Mesh Gateway.

## Import

Service Fabric Mesh Gateway can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_service_fabric_mesh_gateway.gateway1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.ServiceFabricMesh/gateways/gateway1
```
