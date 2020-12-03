package provider

import "github.com/pulumi/pulumi/sdk/v2/go/common/resource"

func isNonEmptyString(value resource.PropertyValue) bool {
	return !value.IsNull() && value.IsString() && (value.StringValue() != "")
}
