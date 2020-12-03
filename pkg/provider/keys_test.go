package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
	"github.com/stvp/assert"
)

func TestKeyCheck(t *testing.T) {
	tests := map[string]struct {
		news         resource.PropertyMap
		wantFailures []*rpc.CheckFailure
	}{
		"nulls": {
			news: resource.PropertyMap{},
			wantFailures: []*rpc.CheckFailure{
				{Property: "name", Reason: "this input must be a non-empty string"},
				{Property: "organizationSlug", Reason: "this input must be a non-empty string"},
				{Property: "projectSlug", Reason: "this input must be a non-empty string"},
			},
		},
		"wrong type": {
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue(1),
				"organizationSlug": resource.NewPropertyValue(1),
				"projectSlug":      resource.NewPropertyValue(1),
			},
			wantFailures: []*rpc.CheckFailure{
				{Property: "name", Reason: "this input must be a non-empty string"},
				{Property: "organizationSlug", Reason: "this input must be a non-empty string"},
				{Property: "projectSlug", Reason: "this input must be a non-empty string"},
			},
		},
		// TODO: slug validation
		//
		// "non-slugs": {
		// 	news: resource.PropertyMap{
		// 		"name":             resource.NewPropertyValue("not a slug"),
		// 		"organizationSlug": resource.NewPropertyValue("not/a/slug"),
		// 		"projectSlug":      resource.NewPropertyValue("not.a.slug"),
		// 	},
		// 	wantFailures: []*rpc.CheckFailure{
		// 		{Property: "name", Reason: "this input must be a slug"},
		// 		{Property: "organizationSlug", Reason: "this input must be a slug"},
		// 		{Property: "projectSlug", Reason: "this input must be a slug"},
		// 	},
		// },
		"correct": {
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue("a name"),
				"organizationSlug": resource.NewPropertyValue("org-slug"),
				"projectSlug":      resource.NewPropertyValue("projectSlug"),
			},
			wantFailures: nil,
		},
	}
	ctx := context.Background()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			prov := sentryProvider{}
			resp, err := prov.keyCheck(ctx, &rpc.CheckRequest{
				News: mustMarshalProperties(tc.news),
			})
			assert.Nil(t, err)
			sort.Sort(byProperty(resp.Failures))
			assert.Equal(t, resp.Failures, tc.wantFailures)
		})
	}
}
