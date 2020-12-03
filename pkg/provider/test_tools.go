package provider

import (
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
	"google.golang.org/protobuf/types/known/structpb"
)

type byProperty []*rpc.CheckFailure

func (s byProperty) Len() int { return len(s) }
func (s byProperty) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byProperty) Less(i, j int) bool {
	return s[i].Property < s[j].Property
}

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
