package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atlassian/go-sentry-api"
	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	logger "github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
)

func (k *sentryProvider) keyCheck(ctx context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
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
	for _, key := range []string{"organizationSlug", "name", "projectSlug"} {
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

	key, err := k.sentryClient.CreateClientKey(sentry.Organization{Slug: &organizationSlug}, sentry.Project{Slug: &projectSlug}, name)
	if err != nil {
		return nil, fmt.Errorf("could not CreateClientKey %v: %v", name, err)
	}
	outputs := outputsFromKey(key)
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

	id := req.GetId()
	organizationSlug, projectSlug, localID, err := parseKeyID(req.GetId())
	if err != nil {
		return nil, err
	}
	key, err := getClientKey(k.sentryClient, sentry.Organization{Slug: &organizationSlug}, sentry.Project{Slug: &projectSlug}, localID)
	if err != nil {
		return nil, err
	}
	if key.ID == "" {
		// The key was not found, delete it from stack state.
		return &rpc.ReadResponse{}, nil
	}
	properties := resource.NewPropertyMapFromMap(outputsFromKey(key))
	state, err := plugin.MarshalProperties(properties, plugin.MarshalOptions{
		Label: label + ".state", KeepUnknowns: true, SkipNulls: true,
	})
	if err != nil {
		return nil, err
	}
	return &rpc.ReadResponse{
		Id:         id,
		Properties: state,
	}, nil
}

func (k *sentryProvider) keyDelete(ctx context.Context, req *rpc.DeleteRequest) (*pbempty.Empty, error) {
	organizationSlug, projectSlug, localID, err := parseKeyID(req.GetId())
	if err != nil {
		return &pbempty.Empty{}, err
	}
	err = k.sentryClient.DeleteClientKey(
		sentry.Organization{Slug: &organizationSlug},
		sentry.Project{Slug: &projectSlug},
		sentry.Key{ID: localID},
	)
	return &pbempty.Empty{}, err
}

func getClientKey(sentryClient sentryClientAPI, organization sentry.Organization, project sentry.Project, localID string) (sentry.Key, error) {
	// The official library does not provide a way to get a single key.  Get
	// the full list and choose the right one.
	keys, err := sentryClient.GetClientKeys(organization, project)
	if err != nil {
		return sentry.Key{}, fmt.Errorf("could not GetClientKeys: %v", err)
	}
	for _, key := range keys {
		if key.ID == localID {
			return key, nil
		}
	}
	return sentry.Key{}, nil
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

func outputsFromKey(key sentry.Key) map[string]interface{} {
	return map[string]interface{}{
		"name":        key.Label,
		"dsnSecret":   key.DSN.Secret,
		"dsnCSP":      key.DSN.CSP,
		"dsnPublic":   key.DSN.Public,
		"secret":      key.Secret,
		"public":      key.Public,
		"dateCreated": key.DateCreated.Format(time.RFC3339),
	}
}
