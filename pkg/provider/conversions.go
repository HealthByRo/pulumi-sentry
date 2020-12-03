package provider

import "github.com/pulumi/pulumi/sdk/v2/go/common/resource"

func stringPtrFromPropertyValue(val resource.PropertyValue) *string {
	if val.IsNull() {
		return nil
	}
	v := val.StringValue()
	return &v
}
