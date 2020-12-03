package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
	"github.com/stvp/assert"
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

func TestProjectCheck(t *testing.T) {
	tests := map[string]struct {
		news         resource.PropertyMap
		wantFailures []*rpc.CheckFailure
	}{
		"nulls": {
			news: resource.PropertyMap{},
			wantFailures: []*rpc.CheckFailure{
				{Property: "name", Reason: "this input must be a non-empty string"},
				{Property: "organizationSlug", Reason: "this input must be a non-empty string"},
				{Property: "slug", Reason: "this input must be a non-empty string"},
				{Property: "teamSlug", Reason: "this input must be a non-empty string"},
			},
		},
		"wrong type": {
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue(1),
				"organizationSlug": resource.NewPropertyValue(1),
				"slug":             resource.NewPropertyValue(1),
				"teamSlug":         resource.NewPropertyValue(1),
			},
			wantFailures: []*rpc.CheckFailure{
				{Property: "name", Reason: "this input must be a non-empty string"},
				{Property: "organizationSlug", Reason: "this input must be a non-empty string"},
				{Property: "slug", Reason: "this input must be a non-empty string"},
				{Property: "teamSlug", Reason: "this input must be a non-empty string"},
			},
		},
		// TODO: slug validation
		//
		// "non-slugs": {
		// 	news: resource.PropertyMap{
		// 		"name":             resource.NewPropertyValue("not a slug"),
		// 		"organizationSlug": resource.NewPropertyValue("not/a/slug"),
		// 		"slug":             resource.NewPropertyValue("not.a.slug"),
		// 		"teamSlug":         resource.NewPropertyValue("not=a=slug"),
		// 	},
		// 	wantFailures: []*rpc.CheckFailure{
		// 		{Property: "name", Reason: "this input must be a slug"},
		// 		{Property: "organizationSlug", Reason: "this input must be a slug"},
		// 		{Property: "slug", Reason: "this input must be a slug"},
		// 		{Property: "teamSlug", Reason: "this input must be a slug"},
		// 	},
		// },
		"correct": {
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue("a name"),
				"organizationSlug": resource.NewPropertyValue("org-slug"),
				"slug":             resource.NewPropertyValue("slug"),
				"teamSlug":         resource.NewPropertyValue("team-slug"),
			},
			wantFailures: nil,
		},
	}
	ctx := context.Background()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			prov := sentryProvider{}
			resp, err := prov.projectCheck(ctx, &rpc.CheckRequest{
				Urn:  "urn:pulumi:fake::fake::fake::fake",
				News: mustMarshalProperties(tc.news),
			})
			assert.Nil(t, err)
			sort.Sort(byProperty(resp.Failures))
			assert.Equal(t, resp.Failures, tc.wantFailures)
		})
	}
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
