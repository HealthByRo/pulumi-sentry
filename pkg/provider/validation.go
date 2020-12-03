package provider

import (
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
)

func checkOptionalString(failures *[]*rpc.CheckFailure, props resource.PropertyMap, key string) {
	value := props[resource.PropertyKey(key)]
	if value.IsNull() {
		return
	}

	if !value.IsString() || (value.StringValue() == "") {
		*failures = append(*failures, &rpc.CheckFailure{
			Property: key,
			Reason:   "this input must be a string",
		})
	}
}

func checkNonEmptyString(failures *[]*rpc.CheckFailure, props resource.PropertyMap, key string) {
	value := props[resource.PropertyKey(key)]
	if value.IsNull() || !value.IsString() || (value.StringValue() == "") {
		*failures = append(*failures, &rpc.CheckFailure{
			Property: key,
			Reason:   "this input must be a non-empty string",
		})
	}
}
