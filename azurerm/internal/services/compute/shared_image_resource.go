package compute

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmSharedImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSharedImageCreateUpdate,
		Read:   resourceArmSharedImageRead,
		Update: resourceArmSharedImageCreateUpdate,
		Delete: resourceArmSharedImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				ValidateFunc: validate.SharedImageName,
			},

			"gallery_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SharedImageGalleryName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"os_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.Linux),
					string(compute.Windows),
				}, false),
			},

			"hyper_v_generation": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(compute.HyperVGenerationTypesV1),
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.V1),
					string(compute.V2),
				}, false),
			},

			"identifier": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"publisher": {
							Type:     schema.TypeString,
							Required: true,
						},
						"offer": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sku": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"eula": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"privacy_statement_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"release_note_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmSharedImageCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.GalleryImagesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Shared Image creation.")

	name := d.Get("name").(string)
	galleryName := d.Get("gallery_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	description := d.Get("description").(string)
	hyperVGeneration := d.Get("hyper_v_generation").(string)

	eula := d.Get("eula").(string)
	privacyStatementUri := d.Get("privacy_statement_uri").(string)
	releaseNoteURI := d.Get("release_note_uri").(string)

	osType := d.Get("os_type").(string)
	t := d.Get("tags").(map[string]interface{})

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, galleryName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_shared_image", *existing.ID)
		}
	}

	identifier := expandGalleryImageIdentifier(d)

	image := compute.GalleryImage{
		Location: utils.String(location),
		GalleryImageProperties: &compute.GalleryImageProperties{
			Description:         utils.String(description),
			Eula:                utils.String(eula),
			Identifier:          identifier,
			PrivacyStatementURI: utils.String(privacyStatementUri),
			ReleaseNoteURI:      utils.String(releaseNoteURI),
			OsType:              compute.OperatingSystemTypes(osType),
			OsState:             compute.Generalized,
			HyperVGeneration:    compute.HyperVGeneration(hyperVGeneration),
		},
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, galleryName, name, image)
	if err != nil {
		return fmt.Errorf("Error creating/updating Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation/update of Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, galleryName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Shared Image %q (Gallery %q / Resource Group %q) ID", name, galleryName, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmSharedImageRead(d, meta)
}

func resourceArmSharedImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.GalleryImagesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	galleryName := id.Path["galleries"]
	name := id.Path["images"]

	resp, err := client.Get(ctx, resourceGroup, galleryName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Shared Image %q (Gallery %q / Resource Group %q) was not found - removing from state", name, galleryName, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("gallery_name", galleryName)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.GalleryImageProperties; props != nil {
		d.Set("description", props.Description)
		d.Set("eula", props.Eula)
		d.Set("os_type", string(props.OsType))
		d.Set("hyper_v_generation", string(props.HyperVGeneration))
		d.Set("privacy_statement_uri", props.PrivacyStatementURI)
		d.Set("release_note_uri", props.ReleaseNoteURI)

		flattenedIdentifier := flattenGalleryImageIdentifier(props.Identifier)
		if err := d.Set("identifier", flattenedIdentifier); err != nil {
			return fmt.Errorf("Error setting `identifier`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmSharedImageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.GalleryImagesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	galleryName := id.Path["galleries"]
	name := id.Path["images"]

	future, err := client.Delete(ctx, resourceGroup, galleryName, name)
	if err != nil {
		return fmt.Errorf("deleting Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("failed to wait for deleting Shared Image %q (Gallery %q / Resource Group %q): %+v", name, galleryName, resourceGroup, err)
	}

	log.Printf("[DEBUG] Waiting for Shared Image %q (Gallery %q / Resource Group %q) to be eventually deleted", name, galleryName, resourceGroup)
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{"Exists"},
		Target:                    []string{"NotFound"},
		Refresh:                   sharedImageDeleteStateRefreshFunc(ctx, client, resourceGroup, name, galleryName),
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 10,
		Timeout:                   d.Timeout(schema.TimeoutDelete),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("failed to wait for Shared Image %q (Gallery %q / Resource Group %q) to be deleted: %+v", name, galleryName, resourceGroup, err)
	}

	return nil
}

func sharedImageDeleteStateRefreshFunc(ctx context.Context, client *compute.GalleryImagesClient, resourceGroupName string, imageName string, galleryName string) resource.StateRefreshFunc {
	// The resource Shared Image depends on the resource Shared Image Gallery.
	// Although the delete API returns 404 which means the Shared Image resource has been deleted.
	// Then it tries to immediately delete Shared Image Gallery but it still throws error `Can not delete resource before nested resources are deleted.`
	// In this case we're going to try triggering the Deletion again, in-case it didn't work prior to this attempt.
	// For more details, see related Bug: https://github.com/Azure/azure-sdk-for-go/issues/8314
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, resourceGroupName, galleryName, imageName)
		if err != nil {
			if utils.ResponseWasNotFound(res.Response) {
				return "NotFound", "NotFound", nil
			}

			return nil, "", fmt.Errorf("failed to poll to check if the Shared Image has been deleted: %+v", err)
		}

		return res, "Exists", nil
	}
}

func expandGalleryImageIdentifier(d *schema.ResourceData) *compute.GalleryImageIdentifier {
	vs := d.Get("identifier").([]interface{})
	v := vs[0].(map[string]interface{})

	offer := v["offer"].(string)
	publisher := v["publisher"].(string)
	sku := v["sku"].(string)

	return &compute.GalleryImageIdentifier{
		Sku:       utils.String(sku),
		Publisher: utils.String(publisher),
		Offer:     utils.String(offer),
	}
}

func flattenGalleryImageIdentifier(input *compute.GalleryImageIdentifier) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	result := make(map[string]interface{})

	if input.Offer != nil {
		result["offer"] = *input.Offer
	}

	if input.Publisher != nil {
		result["publisher"] = *input.Publisher
	}

	if input.Sku != nil {
		result["sku"] = *input.Sku
	}

	return []interface{}{result}
}