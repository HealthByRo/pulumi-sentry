package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/atlassian/go-sentry-api"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
	"github.com/stvp/assert"
)

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

func TestProjectDiff(t *testing.T) {
	baseOlds := resource.PropertyMap{
		"name":             resource.NewPropertyValue("base name"),
		"organizationSlug": resource.NewPropertyValue("base-org-slug"),
		"slug":             resource.NewPropertyValue("base-slug"),
		"teamSlug":         resource.NewPropertyValue("base-team-slug"),
	}
	tests := map[string]struct {
		olds, news   resource.PropertyMap
		wantResponse rpc.DiffResponse
	}{
		"no change": {
			olds:         baseOlds,
			news:         baseOlds,
			wantResponse: rpc.DiffResponse{},
		},
		"simple updates": {
			olds: baseOlds,
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue("new name"),
				"organizationSlug": resource.NewPropertyValue("base-org-slug"),
				"slug":             resource.NewPropertyValue("new-slug"),
				"teamSlug":         resource.NewPropertyValue("new-team-slug"),
			},
			wantResponse: rpc.DiffResponse{
				Changes: rpc.DiffResponse_DIFF_SOME,
				Diffs:   []string{"name", "slug", "teamSlug"},
			},
		},
		"replacement": {
			olds: baseOlds,
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue("base name"),
				"organizationSlug": resource.NewPropertyValue("new-org-slug"),
				"slug":             resource.NewPropertyValue("base-slug"),
				"teamSlug":         resource.NewPropertyValue("base-team-slug"),
			},
			wantResponse: rpc.DiffResponse{
				Changes:             rpc.DiffResponse_DIFF_SOME,
				Diffs:               []string{"organizationSlug"},
				Replaces:            []string{"organizationSlug"},
				DeleteBeforeReplace: true,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			prov := sentryProvider{}
			resp, err := prov.projectDiff(tc.olds, tc.news)
			assert.Nil(t, err)
			assert.Equal(t, tc.wantResponse, *resp)
		})
	}
}

func TestProjectCreate(t *testing.T) {
	ctx := context.Background()
	createCalled := false
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			createProject: func(org sentry.Organization, team sentry.Team, name string, slug *string) (sentry.Project, error) {
				assert.Equal(t, *org.Slug, "the-org")
				assert.Equal(t, *team.Slug, "the-team")
				assert.Equal(t, name, "a name")
				assert.Equal(t, *slug, "slug")
				createCalled = true
				return sentry.Project{
					Name: "name-from-create",
					Slug: stringPtr("slug-from-create"),
				}, nil
			},
			getClientKeys: func(o sentry.Organization, p sentry.Project) ([]sentry.Key, error) {
				return []sentry.Key{
					{Label: "Default", DSN: sentry.DSN{Public: "public-dsn"}},
				}, nil
			},
		},
	}
	inputs := resource.PropertyMap{
		"name":             resource.NewPropertyValue("a name"),
		"organizationSlug": resource.NewPropertyValue("the-org"),
		"slug":             resource.NewPropertyValue("slug"),
		"teamSlug":         resource.NewPropertyValue("the-team"),
	}
	resp, err := prov.projectCreate(ctx, &rpc.CreateRequest{}, inputs)
	assert.Nil(t, err)
	assert.True(t, createCalled)
	assert.Equal(t, resp.GetId(), "the-org/slug-from-create")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"name":                      resource.NewPropertyValue("name-from-create"),
		"organizationSlug":          resource.NewPropertyValue("the-org"),
		"slug":                      resource.NewPropertyValue("slug-from-create"),
		"teamSlug":                  resource.NewPropertyValue("the-team"),
		"defaultClientKeyDSNPublic": resource.NewPropertyValue("public-dsn"),
	})
}

func TestProjectRead(t *testing.T) {
	ctx := context.Background()
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			getClientKeys: func(o sentry.Organization, p sentry.Project) ([]sentry.Key, error) {
				return []sentry.Key{
					{Label: "Default", DSN: sentry.DSN{Public: "public-dsn"}},
				}, nil
			},
			getProject: func(org sentry.Organization, projslug string) (sentry.Project, error) {
				assert.Equal(t, *org.Slug, "org-slug")
				assert.Equal(t, projslug, "proj-slug")
				return sentry.Project{
					Name: "name-from-read",
					Slug: stringPtr("slug-from-read"),
					Team: &sentry.Team{
						Slug: stringPtr("team-slug-from-read"),
					},
				}, nil
			},
		},
	}
	resp, err := prov.projectRead(ctx, &rpc.ReadRequest{Id: "org-slug/proj-slug"})
	assert.Nil(t, err)
	assert.Equal(t, resp.GetId(), "org-slug/slug-from-read")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"name":                      resource.NewPropertyValue("name-from-read"),
		"organizationSlug":          resource.NewPropertyValue("org-slug"),
		"slug":                      resource.NewPropertyValue("slug-from-read"),
		"teamSlug":                  resource.NewPropertyValue("team-slug-from-read"),
		"defaultClientKeyDSNPublic": resource.NewPropertyValue("public-dsn"),
	})
}

func TestProjectDelete(t *testing.T) {
	ctx := context.Background()
	deleteCalled := false
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			deleteProject: func(org sentry.Organization, proj sentry.Project) error {
				assert.Equal(t, *org.Slug, "the-org")
				assert.Equal(t, *proj.Slug, "the-proj")
				deleteCalled = true
				return nil
			},
		},
	}
	_, err := prov.projectDelete(ctx, &rpc.DeleteRequest{Id: "the-org/the-proj"})
	assert.Nil(t, err)
	assert.True(t, deleteCalled)
}
