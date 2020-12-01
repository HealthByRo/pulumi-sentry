package provider

import (
	"context"
	"fmt"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
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
		Id:         slug,
		Properties: outputProperties,
	}, nil
}

func (k *sentryProvider) projectRead(ctx context.Context, req *rpc.ReadRequest, inputs resource.PropertyMap) (*rpc.ReadResponse, error) {
	panic(inputs)
}

func (k *sentryProvider) projectDelete(ctx context.Context, req *rpc.DeleteRequest, inputs resource.PropertyMap) (*pbempty.Empty, error) {
	organizationSlug := inputs["organizationSlug"].StringValue()
	organization, err := k.sentryClient.GetOrganization(organizationSlug)
	if err != nil {
		return nil, err
	}
	slug := inputs["slug"].StringValue()
	project, err := k.sentryClient.GetProject(organization, slug)
	if err != nil {
		return nil, err
	}

	err = k.sentryClient.DeleteProject(organization, project)
	return &pbempty.Empty{}, err
}
