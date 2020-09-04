package validate

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// @jackofallops - as of 2020-08 The service introduced strict compliance with the requirement for `_id` when specifying custom indexes
// For context see https://github.com/terraform-providers/terraform-provider-azurerm/issues/8144
func CheckMongo36IndexRequirements(d *schema.ResourceData) (valid bool) {
	match := []string{"_id"}
	if indexes, ok := d.GetOk("index"); ok {
		for _, v := range indexes.(*schema.Set).List() {
			index := v.(map[string]interface{})
			if reflect.DeepEqual(index["keys"], match) {
				valid = true
				break
			}
		}
	}

	return valid
}