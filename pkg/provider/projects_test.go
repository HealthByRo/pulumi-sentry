package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/marcin-ro/go-sentry-api"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
	"github.com/stvp/assert"
)

func TestProjectCheck(t *testing.T) {
	tests := map[string]struct {
		news         resource.PropertyMap
		wantFailures []*rpc.CheckFailure
	}{
		"nulls for required fields": {
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
				"defaultEnvironment": resource.NewPropertyValue(1),
				"name":               resource.NewPropertyValue(1),
				"organizationSlug":   resource.NewPropertyValue(1),
				"slug":               resource.NewPropertyValue(1),
				"subjectPrefix":      resource.NewPropertyValue(1),
				"subjectTemplate":    resource.NewPropertyValue(1),
				"teamSlug":           resource.NewPropertyValue(1),
			},
			wantFailures: []*rpc.CheckFailure{
				{Property: "defaultEnvironment", Reason: "this input must be a string"},
				{Property: "name", Reason: "this input must be a non-empty string"},
				{Property: "organizationSlug", Reason: "this input must be a non-empty string"},
				{Property: "slug", Reason: "this input must be a non-empty string"},
				{Property: "subjectPrefix", Reason: "this input must be a string"},
				{Property: "subjectTemplate", Reason: "this input must be a string"},
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
		"correct minimal": {
			news: resource.PropertyMap{
				"name":             resource.NewPropertyValue("a name"),
				"organizationSlug": resource.NewPropertyValue("org-slug"),
				"slug":             resource.NewPropertyValue("slug"),
				"teamSlug":         resource.NewPropertyValue("team-slug"),
			},
			wantFailures: nil,
		},
		"correct full": {
			news: resource.PropertyMap{
				"defaultEnvironment": resource.NewPropertyValue("env name"),
				"name":               resource.NewPropertyValue("a name"),
				"organizationSlug":   resource.NewPropertyValue("org-slug"),
				"slug":               resource.NewPropertyValue("slug"),
				"subjectPrefix":      resource.NewPropertyValue("subject prefix"),
				"subjectTemplate":    resource.NewPropertyValue("subject template"),
				"teamSlug":           resource.NewPropertyValue("team-slug"),
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
		"defaultEnvironment": resource.NewPropertyValue("base env name"),
		"name":               resource.NewPropertyValue("base name"),
		"organizationSlug":   resource.NewPropertyValue("base-org-slug"),
		"slug":               resource.NewPropertyValue("base-slug"),
		"subjectPrefix":      resource.NewPropertyValue("base subject prefix"),
		"subjectTemplate":    resource.NewPropertyValue("base subject template"),
		"teamSlug":           resource.NewPropertyValue("base-team-slug"),
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
			news: propertyMapWithOverrides(baseOlds, resource.PropertyMap{
				"defaultEnvironment": resource.NewPropertyValue("new env name"),
				"name":               resource.NewPropertyValue("new name"),
				"slug":               resource.NewPropertyValue("new-slug"),
				"subjectPrefix":      resource.NewPropertyValue("new subject prefix"),
				"subjectTemplate":    resource.NewPropertyValue("new subject template"),
				"teamSlug":           resource.NewPropertyValue("new-team-slug"),
			}),
			wantResponse: rpc.DiffResponse{
				Changes: rpc.DiffResponse_DIFF_SOME,
				Diffs:   []string{"defaultEnvironment", "name", "slug", "teamSlug", "subjectPrefix", "subjectTemplate"},
			},
		},
		"replacement": {
			olds: baseOlds,
			news: propertyMapWithOverrides(baseOlds, resource.PropertyMap{
				"organizationSlug": resource.NewPropertyValue("new-org-slug"),
			}),
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
	updateCalled := false
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
			updateProject: func(org sentry.Organization, proj sentry.Project) error {
				assert.True(t, createCalled)
				assert.Equal(t, *proj.DefaultEnvironment, "env name")
				assert.Equal(t, *proj.SubjectPrefix, "subject prefix")
				assert.Equal(t, *proj.SubjectTemplate, "subject template")
				assert.Equal(t, *proj.DefaultEnvironment, "env name")
				updateCalled = true
				return nil
			},
			getClientKeys: func(o sentry.Organization, p sentry.Project) ([]sentry.Key, error) {
				return []sentry.Key{
					{Label: "Default", DSN: sentry.DSN{Public: "public-dsn"}},
				}, nil
			},
		},
	}
	inputs := resource.PropertyMap{
		"defaultEnvironment": resource.NewPropertyValue("env name"),
		"name":               resource.NewPropertyValue("a name"),
		"organizationSlug":   resource.NewPropertyValue("the-org"),
		"slug":               resource.NewPropertyValue("slug"),
		"subjectPrefix":      resource.NewPropertyValue("subject prefix"),
		"subjectTemplate":    resource.NewPropertyValue("subject template"),
		"teamSlug":           resource.NewPropertyValue("the-team"),
	}
	resp, err := prov.projectCreate(ctx, &rpc.CreateRequest{}, inputs)
	assert.Nil(t, err)
	assert.True(t, updateCalled)
	assert.Equal(t, resp.GetId(), "the-org/slug-from-create")
	assert.Equal(t, mustUnmarshalProperties(resp.GetProperties()), resource.PropertyMap{
		"defaultEnvironment":        resource.NewPropertyValue("env name"),
		"name":                      resource.NewPropertyValue("name-from-create"),
		"organizationSlug":          resource.NewPropertyValue("the-org"),
		"slug":                      resource.NewPropertyValue("slug-from-create"),
		"teamSlug":                  resource.NewPropertyValue("the-team"),
		"subjectPrefix":             resource.NewPropertyValue("subject prefix"),
		"subjectTemplate":           resource.NewPropertyValue("subject template"),
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
					DefaultEnvironment: stringPtr("default-env-from-read"),
					Name:               "name-from-read",
					Slug:               stringPtr("slug-from-read"),
					SubjectPrefix:      stringPtr("subject-prefix-from-read"),
					SubjectTemplate:    stringPtr("subject-template-from-read"),
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
		"defaultEnvironment":        resource.NewPropertyValue("default-env-from-read"),
		"defaultClientKeyDSNPublic": resource.NewPropertyValue("public-dsn"),
		"name":                      resource.NewPropertyValue("name-from-read"),
		"organizationSlug":          resource.NewPropertyValue("org-slug"),
		"slug":                      resource.NewPropertyValue("slug-from-read"),
		"subjectPrefix":             resource.NewPropertyValue("subject-prefix-from-read"),
		"subjectTemplate":           resource.NewPropertyValue("subject-template-from-read"),
		"teamSlug":                  resource.NewPropertyValue("team-slug-from-read"),
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
