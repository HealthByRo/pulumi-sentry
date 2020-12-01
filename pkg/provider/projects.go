package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/atlassian/go-sentry-api"
	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	logger "github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
)

func (k *sentryProvider) projectCheck(ctx context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	// TODO: validate the slug
	return &rpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
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

	organization, err := k.sentryClient.GetOrganization(organizationSlug)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve Organization %#v: %v", organizationSlug, err)
	}
	team, err := k.sentryClient.GetTeam(organization, teamSlug)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve Team: %v", err)
	}

	project, err := k.sentryClient.CreateProject(organization, team, name, &slug)
	if err != nil {
		return nil, fmt.Errorf("could not CreateProject %v: %v", slug, err)
	}
	outputs := map[string]interface{}{
		"organizationSlug": organizationSlug,
		"name":             name,
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
		Id:         buildProjectID(organizationSlug, slug),
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
	return &rpc.ReadResponse{Id: id, Properties: state}, nil
}

func (k *sentryProvider) projectDelete(ctx context.Context, req *rpc.DeleteRequest, inputs resource.PropertyMap) (*pbempty.Empty, error) {
	organizationSlug, slug, err := parseProjectID(req.GetId())
	if err != nil {
		return &pbempty.Empty{}, err
	}
	organization, err := k.sentryClient.GetOrganization(organizationSlug)
	if err != nil {
		return nil, err
	}
	project, err := k.sentryClient.GetProject(organization, slug)
	if err != nil {
		return nil, err
	}
	err = k.sentryClient.DeleteProject(organization, project)
	return &pbempty.Empty{}, err
}

func buildProjectID(organizationSlug, slug string) string {
	return fmt.Sprintf("%s/%s", organizationSlug, slug)
}

func parseProjectID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ID: %s", id)
	}
	return parts[0], parts[1], nil
}
