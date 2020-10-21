package servicefabricmesh

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicefabricmesh/mgmt/2018-09-01-preview/servicefabricmesh"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/servicefabricmesh/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmServiceFabricMeshGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmServiceFabricMeshGatewayCreateUpdate,
		Read:   resourceArmServiceFabricMeshGatewayRead,
		Update: resourceArmServiceFabricMeshGatewayCreateUpdate,
		Delete: resourceArmServiceFabricMeshGatewayDelete,
		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ServiceFabricMeshGatewayID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			// Follow casing issue here https://github.com/Azure/azure-rest-api-specs/issues/9330
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"location": azure.SchemaLocation(),

			"source_network": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"endpoint_references": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},
					},
				},
			},

			"destination_network": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"endpoint_references": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},
					},
				},
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmServiceFabricMeshGatewayCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.GatewayClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Service Fabric Mesh Gateway: %+v", err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_service_fabric_mesh_secret", *existing.ID)
		}
	}

	parameters := servicefabricmesh.GatewayResourceDescription{
		GatewayResourceProperties: &servicefabricmesh.GatewayResourceProperties{
			Description:        utils.String(d.Get("description").(string)),
			SourceNetwork:      expandServiceFabricMeshApplicationNetwork(d.Get("source_network").([]interface{})),
			DestinationNetwork: expandServiceFabricMeshApplicationNetwork(d.Get("source_network").([]interface{})),
		},
		Location: utils.String(location),
		Tags:     tags.Expand(t),
	}

	if _, err := client.Create(ctx, resourceGroup, name, parameters); err != nil {
		return fmt.Errorf("creating Service Fabric Mesh Gateway %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("retrieving Service Fabric Mesh Gateway %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("client returned a nil ID for Service Fabric Mesh Gateway %q", name)
	}

	d.SetId(*resp.ID)

	return resourceArmServiceFabricMeshGatewayRead(d, meta)
}

func resourceArmServiceFabricMeshGatewayRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.GatewayClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceFabricMeshGatewayID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Unable to find Service Fabric Mesh Gateway %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading Service Fabric Mesh Gateway: %+v", err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.GatewayResourceProperties; props != nil {
		d.Set("description", props.Description)
		if err := d.Set("source_network", flattenServiceFabricMeshApplicationNetwork(props.SourceNetwork)); err != nil {
			return fmt.Errorf("setting `source_network`: %+v", err)
		}

		if err := d.Set("destination_network", flattenServiceFabricMeshApplicationNetwork(props.DestinationNetwork)); err != nil {
			return fmt.Errorf("setting `destination_network`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmServiceFabricMeshGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.GatewayClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceFabricMeshGatewayID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Service Fabric Mesh Gateway %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	return nil
}

func expandServiceFabricMeshApplicationNetwork(input []interface{}) *servicefabricmesh.NetworkRef {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	attr := input[0].(map[string]interface{})

	endpointRefsInput := attr["endpoint_references"].(*schema.Set).List()
	endpointRefs := make([]servicefabricmesh.EndpointRef, 0)
	for _, endpoint := range endpointRefsInput {
		endpointRefs = append(endpointRefs, servicefabricmesh.EndpointRef{Name: utils.String(endpoint.(string))})
	}

	return &servicefabricmesh.NetworkRef{
		Name:         utils.String(attr["name"].(string)),
		EndpointRefs: &endpointRefs,
	}
}

func flattenServiceFabricMeshApplicationNetwork(input *servicefabricmesh.NetworkRef) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if input == nil {
		return result
	}
	attr := make(map[string]interface{})
	if input.Name != nil {
		attr["name"] = *input.Name
	}
	if input.EndpointRefs != nil {
		result := make([]interface{}, 0)
		for _, ref := range *input.EndpointRefs {
			if ref.Name != nil {
				result = append(result, *ref.Name)
			}
		}
		attr["endpoint_references"] = result
	}

	return result
}
