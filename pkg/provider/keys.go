package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atlassian/go-sentry-api"
	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	logger "github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
)

func (k *sentryProvider) keyCheck(ctx context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	// TODO: validate the slug
	return &rpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (k *sentryProvider) keyDiff(olds, news resource.PropertyMap) (*rpc.DiffResponse, error) {
	// TODO: be more detailed with Diff results, mind DeleteBeforeReplace,
	d := olds.Diff(news)
	if d == nil {
		return &rpc.DiffResponse{}, nil
	}

	changes := rpc.DiffResponse_DIFF_NONE
	var replaces []string
	for _, key := range []resource.PropertyKey{"organizationSlug", "name", "projectSlug"} {
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

func (k *sentryProvider) keyCreate(ctx context.Context, req *rpc.CreateRequest, inputs resource.PropertyMap) (*rpc.CreateResponse, error) {
	organizationSlug := inputs["organizationSlug"].StringValue()
	projectSlug := inputs["projectSlug"].StringValue()
	name := inputs["name"].StringValue()

	organization, err := k.sentryClient.GetOrganization(organizationSlug)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve Organization %#v: %v", organizationSlug, err)
	}
	project, err := k.sentryClient.GetProject(organization, projectSlug)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve Project %#v: %v", projectSlug, err)
	}
	key, err := k.sentryClient.CreateClientKey(organization, project, name)
	if err != nil {
		return nil, fmt.Errorf("could not CreateClientKey %v: %v", name, err)
	}
	outputs := map[string]interface{}{
		"organizationSlug": organizationSlug,
		"name":             name,
		"projectSlug":      projectSlug,
		"dsnSecret":        key.DSN.Secret,
		"dsnCSP":           key.DSN.CSP,
		"dsnPublic":        key.DSN.Public,
		"secret":           key.Secret,
		"public":           key.Public,
		"dateCreated":      key.DateCreated.Format(time.RFC3339),
	}
	outputProperties, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputs),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	return &rpc.CreateResponse{
		Id:         buildKeyID(organizationSlug, projectSlug, key.ID),
		Properties: outputProperties,
	}, nil
}

func (k *sentryProvider) keyRead(ctx context.Context, req *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Read(%s)", k.label(), urn)
	logger.V(9).Infof("%s executing", label)
	panic("not implemented")
}

func (k *sentryProvider) keyDelete(ctx context.Context, req *rpc.DeleteRequest, inputs resource.PropertyMap) (*pbempty.Empty, error) {
	organizationSlug, projectSlug, localID, err := parseKeyID(req.GetId())
	if err != nil {
		return &pbempty.Empty{}, err
	}
	organization, err := k.sentryClient.GetOrganization(organizationSlug)
	if err != nil {
		return nil, err
	}
	project, err := k.sentryClient.GetProject(organization, projectSlug)
	if err != nil {
		return nil, err
	}
	key, err := getClientKey(k.sentryClient, organization, project, localID)
	if err != nil {
		return nil, err
	}
	err = k.sentryClient.DeleteClientKey(organization, project, key)
	return &pbempty.Empty{}, err
}

func getClientKey(sentryClient *sentry.Client, organization sentry.Organization, project sentry.Project, localID string) (sentry.Key, error) {
	// This looks awkward, but the official library does not provide a way to
	// get a single key.
	keys, err := sentryClient.GetClientKeys(organization, project)
	if err != nil {
		return sentry.Key{}, fmt.Errorf("could not GetClientKeys: %v", err)
	}
	for _, key := range keys {
		if key.ID == localID {
			return key, nil
		}
	}
	return sentry.Key{}, fmt.Errorf("could not find a ClientKey matching %#v", localID)
}

func buildKeyID(organizationSlug, slug, localID string) string {
	return fmt.Sprintf("%s/%s/%s", organizationSlug, slug, localID)
}

func parseKeyID(id string) (organizationSlug, projectSlug, localID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid ID: %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}
