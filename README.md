> **:warning: This fork is not currently maintained and will be archived then removed in the future.**

# Pulumi Provider for Sentry

This repository implements a Pulumi provider for Sentry resources. It started
as a clone of the [Pulumi provider boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate).

**THIS IS EARLY STAGE, WORK IN PROGRESS CODE**. Not complete for real world usage yet.

## Overview

Most of the contents of this repo are generated by `make rebuild-sdk`; most of the real, non-automatic work is:

- `schema.json`: the definition of resources published by the provider; any
  changes to this file require rebuilding all the SDKs, see `make rebuild-sdk`,
- `pkg/provider/*`: the actual provider implementation; if you change it you
  need to rebuild and install the provider binary, see `make go-install-provider`,
- `examples/sample-project`: a test project for this provider.

For simplicity I iterate on this project by calling variations on:

```
make rebuild-sdk install-provider && pulumi -C examples/sample-project/ up
```

## Running the sample project

1. Get a sentry account somewhere; free accounts on sentry.io are good enough.
   Create a team, it will be a bit easier if you use the name `test-team` for it.
2. In Sentry, "Organization Settings" -> "Developer Settings" create an
   "Internal integration" and make sure it has Admin permissions to Project,
   Team, and Organization resources.
3. Generate a token for this integration.

Create a new stack with:

```
cd examples/sample-app
pulumi stack init test
pulumi config set sentry:apiURL https://sentry.io/api/0/
pulumi config set --secret sentry:token <integration-token>
```

You might need to find the right API URL if you use Sentry other than https://sentry.io.

To make testing easier, `sample-project` is configurable via environment
variables. You will have to override at least the organization slug, see
`examples/sample-project/main.go` for the list of variables. You can also test
adding or removing the project by setting `SKIP_PROJECT=1`.

## References

Other resoruces for learning about the Pulumi resource model:

- [Pulumi provider boilerplate](https://github.com/pulumi/pulumi-provider-boilerplate)
- [Pulumi Kubernetes provider](https://github.com/pulumi/pulumi-kubernetes/blob/master/provider/pkg/provider/provider.go)
- [Pulumi Terraform Remote State provider](https://github.com/pulumi/pulumi-terraform/blob/master/provider/cmd/pulumi-resource-terraform/provider.go)
