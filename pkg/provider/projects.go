package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/atlassian/go-sentry-api"
	pbempty "github.com/golang/protobuf/ptypes/empty"
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
	for _, key := range []string{"organizationSlug", "name", "slug", "teamSlug"} {
		value := news[resource.PropertyKey(key)]
		if !isNonEmptyString(value) {
			failures = append(failures, &rpc.CheckFailure{
				Property: key,
				Reason:   "this input must be a non-empty string",
			})
			continue
		}
	}

	return &rpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (k *sentryProvider) projectDiff(olds, news resource.PropertyMap) (*rpc.DiffResponse, error) {
	// TODO: be more detailed with Diff results, mind DeleteBeforeReplace,

	d := olds.Diff(news)
	if d == nil {
		return &rpc.DiffResponse{}, nil
	}

	changes := rpc.DiffResponse_DIFF_NONE
	var replaces []string
	for _, key := range []resource.PropertyKey{"organizationSlug", "name", "slug", "teamSlug"} {
		if d.Changed(key) {
			changes = rpc.DiffResponse_DIFF_SOME
			replaces = append(replaces, string(key))
		}
	}

	return &rpc.DiffResponse{
		Changes:  changes,
		Replaces: replaces,
	}, nil
}

func (k *sentryProvider) projectCreate(ctx context.Context, req *rpc.CreateRequest, inputs resource.PropertyMap) (*rpc.CreateResponse, error) {
	organizationSlug := inputs["organizationSlug"].StringValue()
	name := inputs["name"].StringValue()
	slug := inputs["slug"].StringValue()
	teamSlug := inputs["teamSlug"].StringValue()

	project, err := k.sentryClient.CreateProject(sentry.Organization{Slug: &organizationSlug}, sentry.Team{Slug: &teamSlug}, name, &slug)
	if err != nil {
		return nil, fmt.Errorf("could not CreateProject %v: %v", slug, err)
	}
	outputs := map[string]interface{}{
		"organizationSlug": organizationSlug,
		"name":             project.Name,
		"slug":             *project.Slug,
		"teamSlug":         teamSlug,
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

func (k *sentryProvider) projectRead(ctx context.Context, req *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Read(%s)", k.label(), urn)
	logger.V(9).Infof("%s executing", label)

	id := req.GetId()
	organizationSlug, slug, err := parseProjectID(id)
	if err != nil {
		return nil, err
	}
	organization, err := k.sentryClient.GetOrganization(organizationSlug)
	if err != nil {
		return nil, err
	}
	project, err := k.sentryClient.GetProject(organization, slug)
	if err != nil {
		if apiError, ok := err.(*sentry.APIError); ok {
			if apiError.StatusCode == 404 {
				// The project is not there, delete it from stack state.
				return &rpc.ReadResponse{}, nil
			}
			// All other errors: just report them.
		}
		return nil, err
	}
	properties := resource.NewPropertyMapFromMap(map[string]interface{}{
		"organizationSlug": *project.Organization.Slug,
		"name":             project.Name,
		"slug":             *project.Slug,
		"teamSlug":         *project.Team.Slug,
	})
	state, err := plugin.MarshalProperties(properties, plugin.MarshalOptions{
		Label: label + ".state", KeepUnknowns: true, SkipNulls: true,
	})
	if err != nil {
		return nil, err
	}
	return &rpc.ReadResponse{
		Id:         buildProjectID(*project.Organization.Slug, *project.Slug),
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
