package provider

import (
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

func mustMarshalProperties(props resource.PropertyMap) *structpb.Struct {
	marshaled, err := plugin.MarshalProperties(props, plugin.MarshalOptions{
		Label:         "test-label",
		KeepUnknowns:  true,
		KeepSecrets:   true,
		KeepResources: true,
	})
	if err != nil {
		panic(err)
	}
	return marshaled
}

func mustUnmarshalProperties(mprops *structpb.Struct) resource.PropertyMap {
	props, err := plugin.UnmarshalProperties(mprops, plugin.MarshalOptions{
		KeepUnknowns:  true,
		KeepSecrets:   true,
		KeepResources: true,
	})
	if err != nil {
		panic(err)
	}
	return props
}

func stringPtr(value string) *string {
	return &value
}
