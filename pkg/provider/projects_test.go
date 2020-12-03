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

func TestProjectCreate(t *testing.T) {
	ctx := context.Background()
	createCalled := false
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			getOrganization: func(orgSlug string) (sentry.Organization, error) {
				return sentry.Organization{Slug: stringPtr("slug-from-getOrganization")}, nil
			},
			getTeam: func(org sentry.Organization, teamSlug string) (sentry.Team, error) {
				return sentry.Team{Slug: stringPtr("slug-from-getTeam")}, nil
			},
			createProject: func(org sentry.Organization, team sentry.Team, name string, slug *string) (sentry.Project, error) {
				assert.Equal(t, *org.Slug, "slug-from-getOrganization")
				assert.Equal(t, *team.Slug, "slug-from-getTeam")
				assert.Equal(t, name, "a name")
				assert.Equal(t, *slug, "slug")
				createCalled = true
				return sentry.Project{
					Name: "name-from-fake-sentry",
					Slug: stringPtr("slug-from-fake-sentry"),
				}, nil
			},
		},
	}
	inputs := resource.PropertyMap{
		"name":             resource.NewPropertyValue("a name"),
		"organizationSlug": resource.NewPropertyValue("org-slug"),
		"slug":             resource.NewPropertyValue("slug"),
		"teamSlug":         resource.NewPropertyValue("team-slug"),
	}
	resp, err := prov.projectCreate(ctx, &rpc.CreateRequest{}, inputs)
	assert.Nil(t, err)
	assert.True(t, createCalled)
	assert.Equal(t, resp.GetId(), "slug-from-getOrganization/slug-from-fake-sentry")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"name":             resource.NewPropertyValue("name-from-fake-sentry"),
		"organizationSlug": resource.NewPropertyValue("slug-from-getOrganization"),
		"slug":             resource.NewPropertyValue("slug-from-fake-sentry"),
		"teamSlug":         resource.NewPropertyValue("slug-from-getTeam"),
	})
}

func TestProjectRead(t *testing.T) {
	ctx := context.Background()
	prov := sentryProvider{
		sentryClient: &sentryClientMock{
			getOrganization: func(orgSlug string) (sentry.Organization, error) {
				return sentry.Organization{Slug: &orgSlug, Name: "org " + orgSlug}, nil
			},
			getProject: func(org sentry.Organization, projslug string) (sentry.Project, error) {
				assert.Equal(t, *org.Slug, "org-slug")
				assert.Equal(t, projslug, "proj-slug")
				return sentry.Project{
					Name: "name-from-read",
					Slug: stringPtr("slug-from-read"),
					Organization: &sentry.Organization{
						Name: "the org from read",
						Slug: stringPtr("the-org-from-read"),
					},
					Team: &sentry.Team{
						Name: "the team from read",
						Slug: stringPtr("the-team-from-read"),
					},
				}, nil
			},
		},
	}
	resp, err := prov.projectRead(ctx, &rpc.ReadRequest{Id: "org-slug/proj-slug"})
	assert.Nil(t, err)
	assert.Equal(t, resp.GetId(), "the-org-from-read/slug-from-read")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"name":             resource.NewPropertyValue("name-from-read"),
		"organizationSlug": resource.NewPropertyValue("the-org-from-read"),
		"slug":             resource.NewPropertyValue("slug-from-read"),
		"teamSlug":         resource.NewPropertyValue("the-team-from-read"),
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
