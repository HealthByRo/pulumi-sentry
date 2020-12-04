package provider

import (
	"context"
	"fmt"
	"strings"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/marcin-ro/go-sentry-api"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	logger "github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
)

func (k *sentryProvider) projectCheck(ctx context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Check(%s)", k.label(), urn)
	logger.V(9).Infof("%s executing", label)

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.news", label),
		KeepUnknowns: true,
		SkipNulls:    true,
		RejectAssets: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "check failed because malformed resource inputs")
	}

	var failures []*rpc.CheckFailure
	checkOptionalString(&failures, news, "defaultEnvironment")
	checkNonEmptyString(&failures, news, "organizationSlug")
	checkNonEmptyString(&failures, news, "name")
	checkNonEmptyString(&failures, news, "slug")
	checkOptionalString(&failures, news, "subjectPrefix")
	checkOptionalString(&failures, news, "subjectTemplate")
	checkNonEmptyString(&failures, news, "teamSlug")

	return &rpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (k *sentryProvider) projectDiff(olds, news resource.PropertyMap) (*rpc.DiffResponse, error) {
	d := olds.Diff(news)
	if d == nil {
		return &rpc.DiffResponse{}, nil
	}

	changeRequiresReplacement := map[string]bool{
		// Organization and project slugs are part of the Project's ID.
		//
		// Sentry API allows changing the project slug, but Pulumi does not
		// allow updating a resource ID in `update`, so if we were to allow
		// that we would have to use artificial IDs on Pulumi side instead of
		// the fully readable and predictable <orgSlug>/<projSlug>.
		//
		// It does not sound worth it, let's just assume that changing a
		// project's slug requires a replacement.
		"organizationSlug": true,
		"slug":             true,

		// Changing a project's owner team is a more complex operation now: I
		// can't find any up to date documentation about it, only deprecation
		// warnings on updating the team via the "update project" endpoint (it
		// is already deprecated and not functioning on sentry.io).  I checked
		// in Sentry code and it was already mostly refactored to not have the
		// notion of an "owner team", just "teams with access to" (although it
		// pretends to have one of them singled out as "the team" for the
		// project).
		//
		// Taking that into account I'm shelving this.  Doing it cleanly will
		// require changing the input on the Project resource from a single
		// teamSlug to an array of slugs, and updating the go-sentry-api
		// library, all so that we don't spend time handling the old
		// abstraction that Sentry is already out of.  For the time being let's
		// handle team changes by replacements and return to them when we know
		// we need them.
		"teamSlug": true,
	}
	var diffs, replaces []string
	for _, key := range d.Keys() {
		if d.Changed(key) {
			diffs = append(diffs, string(key))
			if changeRequiresReplacement[string(key)] {
				replaces = append(replaces, string(key))
			}
		}
	}

	changes := rpc.DiffResponse_DIFF_NONE
	if len(diffs) > 0 {
		changes = rpc.DiffResponse_DIFF_SOME
	}

	return &rpc.DiffResponse{
		Changes:             changes,
		Diffs:               diffs,
		Replaces:            replaces,
		DeleteBeforeReplace: len(replaces) > 0,
	}, nil
}

func (k *sentryProvider) projectCreate(ctx context.Context, req *rpc.CreateRequest, inputs resource.PropertyMap) (*rpc.CreateResponse, error) {
	organizationSlug := inputs["organizationSlug"].StringValue()
	name := inputs["name"].StringValue()
	slug := inputs["slug"].StringValue()
	teamSlug := inputs["teamSlug"].StringValue()

	org := sentry.Organization{Slug: &organizationSlug}
	project, err := k.sentryClient.CreateProject(org, sentry.Team{Slug: &teamSlug}, name, &slug)
	if err != nil {
		return nil, fmt.Errorf("could not CreateProject %v: %v", slug, err)
	}

	project.DefaultEnvironment = stringPtrFromPropertyValue(inputs["defaultEnvironment"])
	project.SubjectPrefix = stringPtrFromPropertyValue(inputs["subjectPrefix"])
	project.SubjectTemplate = stringPtrFromPropertyValue(inputs["subjectTemplate"])

	if err := k.sentryClient.UpdateProject(org, project); err != nil {
		return nil, fmt.Errorf("could not UpdateProject %v: %v", project.Slug, err)
	}

	defaultKey, err := getDefaultClientKey(k.sentryClient, organizationSlug, slug)
	if err != nil {
		return nil, fmt.Errorf("could not get default ClientKey for %v: %v", slug, err)
	}

	outputs := map[string]interface{}{
		"defaultClientKeyDSNPublic": defaultKey.DSN.Public,
		"defaultEnvironment":        project.DefaultEnvironment,
		"name":                      project.Name,
		"organizationSlug":          organizationSlug,
		"slug":                      *project.Slug,
		"subjectPrefix":             project.SubjectPrefix,
		"subjectTemplate":           project.SubjectTemplate,
		"teamSlug":                  teamSlug,
	}

	outputProperties, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputs),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	return &rpc.CreateResponse{
		Id:         buildProjectID(organizationSlug, *project.Slug),
		Properties: outputProperties,
	}, nil
}

func (k *sentryProvider) projectUpdate(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Read(%s)", k.label(), urn)
	logger.V(9).Infof("%s executing", label)

	organizationSlug, slug, err := parseProjectID(req.GetId())
	if err != nil {
		return nil, err
	}
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.olds", label),
	})
	if err != nil {
		return nil, fmt.Errorf("failed projectUpdate because of malformed resource inputs: %w", err)
	}
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.news", label),
	})
	if err != nil {
		return nil, fmt.Errorf("failed projectUpdate because of malformed resource inputs: %w", err)
	}

	d := olds.Diff(news)
	if d == nil {
		// This would be really surprising, pulumi should not let that happen.
		return &rpc.UpdateResponse{}, nil
	}

	// This should already be validated in Diff, but let's be strict just in case.
	allowChanges := map[string]bool{
		"defaultEnvironment": true,
		"name":               true,
		"subjectPrefix":      true,
		"subjectTemplate":    true,
	}
	for _, key := range d.Keys() {
		if d.Changed(key) && !allowChanges[string(key)] {
			return nil, fmt.Errorf("projectUpdate: don't know how to change %v", key)
		}
	}

	project := sentry.Project{
		Slug:               &slug,
		DefaultEnvironment: stringPtr(news["defaultEnvironment"].StringValue()),
		Name:               news["name"].StringValue(),
		SubjectPrefix:      stringPtr(news["subjectPrefix"].StringValue()),
		SubjectTemplate:    stringPtr(news["subjectTemplate"].StringValue()),
	}

	err = k.sentryClient.UpdateProject(sentry.Organization{Slug: &organizationSlug}, project)
	return &rpc.UpdateResponse{Properties: req.GetNews()}, nil
}

func (k *sentryProvider) projectRead(ctx context.Context, req *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Read(%s)", k.label(), urn)
	logger.V(9).Infof("%s executing", label)

	id := req.GetId()
	organizationSlug, slug, err := parseProjectID(id)
	if err != nil {
		return nil, err
	}
	project, err := k.sentryClient.GetProject(sentry.Organization{Slug: &organizationSlug}, slug)
	if err != nil {
		if apiError, ok := err.(sentry.APIError); ok {
			if apiError.StatusCode == 404 {
				// The project is not there, delete it from stack state.
				return &rpc.ReadResponse{}, nil
			}
			// All other errors: just report them.
		}
		return nil, err
	}
	defaultKey, err := getDefaultClientKey(k.sentryClient, organizationSlug, slug)
	if err != nil {
		return nil, fmt.Errorf("could not get default ClientKey for %v: %v", slug, err)
	}
	properties := resource.NewPropertyMapFromMap(map[string]interface{}{
		"defaultClientKeyDSNPublic": defaultKey.DSN.Public,
		"defaultEnvironment":        project.DefaultEnvironment,
		"organizationSlug":          organizationSlug,
		"name":                      project.Name,
		"slug":                      *project.Slug,
		"subjectPrefix":             project.SubjectPrefix,
		"subjectTemplate":           project.SubjectTemplate,
		"teamSlug":                  *project.Team.Slug,
	})
	state, err := plugin.MarshalProperties(properties, plugin.MarshalOptions{
		Label: label + ".state", KeepUnknowns: true, SkipNulls: true,
	})
	if err != nil {
		return nil, err
	}
	return &rpc.ReadResponse{
		Id:         buildProjectID(organizationSlug, *project.Slug),
		Properties: state,
	}, nil
}

func (k *sentryProvider) projectDelete(ctx context.Context, req *rpc.DeleteRequest) (*pbempty.Empty, error) {
	organizationSlug, slug, err := parseProjectID(req.GetId())
	if err != nil {
		return &pbempty.Empty{}, err
	}
	err = k.sentryClient.DeleteProject(sentry.Organization{Slug: &organizationSlug}, sentry.Project{Slug: &slug})
	return &pbempty.Empty{}, err
}

func buildProjectID(organizationSlug, slug string) string {
	return fmt.Sprintf("%s/%s", organizationSlug, slug)
}

func parseProjectID(id string) (organizationSlug, slug string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ID: %s", id)
	}
	return parts[0], parts[1], nil
}

func getDefaultClientKey(sentryClient sentryClientAPI, organizationSlug, slug string) (sentry.Key, error) {
	keys, err := sentryClient.GetClientKeys(
		sentry.Organization{Slug: &organizationSlug},
		sentry.Project{Slug: &slug},
	)
	if err != nil {
		return sentry.Key{}, err
	}

	for _, key := range keys {
		if key.Label == "Default" {
			return key, nil
		}
	}

	return sentry.Key{}, nil
}
