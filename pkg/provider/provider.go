// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"fmt"

	"github.com/atlassian/go-sentry-api"
	"github.com/pulumi/pulumi/pkg/v2/resource/provider"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	logger "github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	rpc "github.com/pulumi/pulumi/sdk/v2/proto/go"

	pbempty "github.com/golang/protobuf/ptypes/empty"
)

type sentryProvider struct {
	host    *provider.HostClient
	name    string
	version string

	sentryClient sentryClientAPI
}

func makeProvider(host *provider.HostClient, name, version string) (rpc.ResourceProviderServer, error) {
	// Return the new provider
	return &sentryProvider{
		host:    host,
		name:    name,
		version: version,
	}, nil
}

// CheckConfig validates the configuration for this provider.
func (k *sentryProvider) CheckConfig(ctx context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	return &rpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig diffs the configuration for this provider.
func (k *sentryProvider) DiffConfig(ctx context.Context, req *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.DiffConfig(%s)", k.label(), urn)
	logger.V(9).Infof("%s executing", label)

	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.olds", label),
	})
	if err != nil {
		return nil, err
	}
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.news", label),
	})
	if err != nil {
		return nil, fmt.Errorf("failed DiffConfig because of malformed resource inputs: %w", err)
	}

	var diffs []string
	if olds["sentryToken"] != news["sentryToken"] {
		diffs = append(diffs, "sentryToken")
	}
	if olds["sentryApiURL"] != news["sentryApiURL"] {
		diffs = append(diffs, "sentryApiURL")
	}
	if len(diffs) > 0 {
		return &rpc.DiffResponse{
			Changes: rpc.DiffResponse_DIFF_SOME,
			Diffs:   diffs,
		}, nil
	}

	return &rpc.DiffResponse{}, nil
}

// Configure configures the resource provider with "globals" that control its behavior.
func (k *sentryProvider) Configure(ctx context.Context, req *rpc.ConfigureRequest) (*rpc.ConfigureResponse, error) {
	vars := req.GetVariables()
	logger.V(9).Infof("vars %v", vars)

	apiURL := vars["sentry:config:apiURL"]
	var err error
	k.sentryClient, err = sentry.NewClient(vars["sentry:config:token"], &apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize a sentry API client: %v", err)
	}

	return &rpc.ConfigureResponse{}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (k *sentryProvider) Invoke(_ context.Context, req *rpc.InvokeRequest) (*rpc.InvokeResponse, error) {
	tok := req.GetTok()
	return nil, fmt.Errorf("Unknown Invoke token '%s'", tok)
}

// StreamInvoke dynamically executes a built-in function in the provider. The result is streamed
// back as a series of messages.
func (k *sentryProvider) StreamInvoke(req *rpc.InvokeRequest, server rpc.ResourceProvider_StreamInvokeServer) error {
	tok := req.GetTok()
	return fmt.Errorf("Unknown StreamInvoke token '%s'", tok)
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (k *sentryProvider) Check(ctx context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	urn := resource.URN(req.GetUrn())
	ty := urn.Type()
	switch ty {
	case "sentry:index:Project":
		return k.projectCheck(ctx, req)
	case "sentry:index:ClientKey":
		return k.keyCheck(ctx, req)
	}
	return nil, fmt.Errorf("Unknown resource type '%s'", ty)
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (k *sentryProvider) Diff(ctx context.Context, req *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	urn := resource.URN(req.GetUrn())
	ty := urn.Type()

	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	switch ty {
	case "sentry:index:Project":
		return k.projectDiff(olds, news)
	case "sentry:index:ClientKey":
		return k.keyDiff(olds, news)
	}

	return nil, fmt.Errorf("Unknown resource type '%s'", ty)
}

// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
func (k *sentryProvider) Create(ctx context.Context, req *rpc.CreateRequest) (*rpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	urn := resource.URN(req.GetUrn())
	ty := urn.Type()
	switch ty {
	case "sentry:index:Project":
		return k.projectCreate(ctx, req, inputs)
	case "sentry:index:ClientKey":
		return k.keyCreate(ctx, req, inputs)
	}
	return nil, fmt.Errorf("Unknown resource type '%s'", ty)
}

// Read the current live state associated with a resource.
func (k *sentryProvider) Read(ctx context.Context, req *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	urn := resource.URN(req.GetUrn())
	ty := urn.Type()
	switch ty {
	case "sentry:index:Project":
		return k.projectRead(ctx, req)
	case "sentry:index:ClientKey":
		return k.keyRead(ctx, req)
	}

	return nil, fmt.Errorf("Unknown resource type '%s'", ty)
}

// Update updates an existing resource with new values.
func (k *sentryProvider) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	urn := resource.URN(req.GetUrn())
	ty := urn.Type()

	switch ty {
	case "sentry:index:Project":
		panic("Update not implemented for sentry:index:Project")
	case "sentry:index:ClientKey":
		panic("Update not implemented for sentry:index:ClientKey")
	}

	return nil, fmt.Errorf("Unknown resource type '%s'", ty)
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (k *sentryProvider) Delete(ctx context.Context, req *rpc.DeleteRequest) (*pbempty.Empty, error) {
	urn := resource.URN(req.GetUrn())
	ty := urn.Type()
	switch ty {
	case "sentry:index:Project":
		return k.projectDelete(ctx, req)
	case "sentry:index:ClientKey":
		return k.keyDelete(ctx, req)
	}
	return nil, fmt.Errorf("Unknown resource type '%s'", ty)

}

// Construct creates a new component resource.
func (k *sentryProvider) Construct(_ context.Context, _ *rpc.ConstructRequest) (*rpc.ConstructResponse, error) {
	panic("Construct not implemented")
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (k *sentryProvider) GetPluginInfo(context.Context, *pbempty.Empty) (*rpc.PluginInfo, error) {
	return &rpc.PluginInfo{
		Version: k.version,
	}, nil
}

// GetSchema returns the JSON-serialized schema for the provider.
func (k *sentryProvider) GetSchema(ctx context.Context, req *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error) {
	return &rpc.GetSchemaResponse{}, nil
}

// Cancel signals the provider to gracefully shut down and abort any ongoing resource operations.
// Operations aborted in this way will return an error (e.g., `Update` and `Create` will either a
// creation error or an initialization error). Since Cancel is advisory and non-blocking, it is up
// to the host to decide how long to wait after Cancel is called before (e.g.)
// hard-closing any gRPC connection.
func (k *sentryProvider) Cancel(context.Context, *pbempty.Empty) (*pbempty.Empty, error) {
	// TODO
	return &pbempty.Empty{}, nil
}

func (k *sentryProvider) label() string {
	return fmt.Sprintf("Provider[%s]", k.name)
}
