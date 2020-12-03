package provider

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/atlassian/go-sentry-api"
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

func TestKeyCreate(t *testing.T) {
	ctx := context.Background()
	createCalled := false
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			createClientKey: func(org sentry.Organization, project sentry.Project, name string) (sentry.Key, error) {
				assert.Equal(t, *org.Slug, "org-slug")
				assert.Equal(t, *project.Slug, "proj-slug")
				assert.Equal(t, name, "key name")
				createCalled = true
				return sentry.Key{
					Label: "label-from-create",
					DSN: sentry.DSN{
						Secret: "dsn-secret-from-create",
						CSP:    "dsn-csp-from-create",
						Public: "dsn-public-from-create",
					},
					Secret:      "secret-from-create",
					ID:          "id-from-create",
					DateCreated: time.Date(2020, 12, 31, 12, 34, 56, 0, time.UTC),
					Public:      "public-from-create",
				}, nil
			},
		},
	}
	inputs := resource.PropertyMap{
		"name":             resource.NewPropertyValue("key name"),
		"organizationSlug": resource.NewPropertyValue("org-slug"),
		"projectSlug":      resource.NewPropertyValue("proj-slug"),
	}
	resp, err := prov.keyCreate(ctx, &rpc.CreateRequest{}, inputs)
	assert.Nil(t, err)
	assert.True(t, createCalled)
	assert.Equal(t, resp.GetId(), "org-slug/proj-slug/id-from-create")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"name":        resource.NewPropertyValue("label-from-create"),
		"dsnSecret":   resource.NewPropertyValue("dsn-secret-from-create"),
		"dsnCSP":      resource.NewPropertyValue("dsn-csp-from-create"),
		"dsnPublic":   resource.NewPropertyValue("dsn-public-from-create"),
		"secret":      resource.NewPropertyValue("secret-from-create"),
		"public":      resource.NewPropertyValue("public-from-create"),
		"dateCreated": resource.NewPropertyValue("2020-12-31T12:34:56Z"),
	})
}

func TestKeyRead(t *testing.T) {
	ctx := context.Background()
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			getClientKeys: func(org sentry.Organization, proj sentry.Project) ([]sentry.Key, error) {
				assert.Equal(t, *org.Slug, "org-slug")
				assert.Equal(t, *proj.Slug, "proj-slug")

				return []sentry.Key{
					{
						Label: "label-1",
						DSN: sentry.DSN{
							Secret: "dsn-secret-1",
							CSP:    "dsn-csp-1",
							Public: "dsn-public-1",
						},
						Secret:      "secret-1",
						ID:          "id-1",
						DateCreated: time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC),
						Public:      "public-1",
					},
					{
						Label: "label-2",
						DSN: sentry.DSN{
							Secret: "dsn-secret-2",
							CSP:    "dsn-csp-2",
							Public: "dsn-public-2",
						},
						Secret:      "secret-2",
						ID:          "id-2",
						DateCreated: time.Date(2020, 2, 2, 2, 2, 2, 0, time.UTC),
						Public:      "public-2",
					},
				}, nil
			},
		},
	}
	resp, err := prov.keyRead(ctx, &rpc.ReadRequest{Id: "org-slug/proj-slug/id-2"})
	assert.Nil(t, err)
	assert.Equal(t, resp.GetId(), "org-slug/proj-slug/id-2")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"name":        resource.NewPropertyValue("label-2"),
		"dsnSecret":   resource.NewPropertyValue("dsn-secret-2"),
		"dsnCSP":      resource.NewPropertyValue("dsn-csp-2"),
		"dsnPublic":   resource.NewPropertyValue("dsn-public-2"),
		"secret":      resource.NewPropertyValue("secret-2"),
		"public":      resource.NewPropertyValue("public-2"),
		"dateCreated": resource.NewPropertyValue("2020-02-02T02:02:02Z"),
	})
}

func TestKeyDelete(t *testing.T) {
	ctx := context.Background()
	deleteCalled := false
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			deleteClientKey: func(org sentry.Organization, proj sentry.Project, key sentry.Key) error {
				assert.Equal(t, *org.Slug, "the-org")
				assert.Equal(t, *proj.Slug, "the-proj")
				assert.Equal(t, key.ID, "keyID")
				deleteCalled = true
				return nil
			},
		},
	}
	_, err := prov.keyDelete(ctx, &rpc.DeleteRequest{Id: "the-org/the-proj/keyID"})
	assert.Nil(t, err)
	assert.True(t, deleteCalled)
}
