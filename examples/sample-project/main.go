package main

import (
	"os"

	"github.com/HealthByRo/pulumi-sentry/sdk/go/sentry"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func main() {
	pulumi.Run(createProjects)
}

func createProjects(ctx *pulumi.Context) error {
	// This code may look a bit weird but it's used solely for manual
	// integration testing and experiments, and all the environment variables
	// allow us to introduce variability between runs without having to change
	// the code.
	skipProject := os.Getenv("SKIP_PROJECT") != ""
	orgSlug := os.Getenv("ORG_SLUG")
	if orgSlug == "" {
		panic("You must provide ORG_SLUG env variable")
	}

	if !skipProject {
		projectOutput, err := sentry.NewProject(ctx, "testing", &sentry.ProjectArgs{
			Name:             pulumi.String(getenvWithDefault("PROJ_NAME", "Sample Project")),
			Slug:             pulumi.String(getenvWithDefault("PROJ_SLUG", "sample-project")),
			OrganizationSlug: pulumi.String(orgSlug),
			TeamSlug:         pulumi.String(getenvWithDefault("TEAM_SLUG", "test-team")),
		})
		if err != nil {
			return err
		}
		ctx.Export("sentry-project-name", projectOutput.Name)
		ctx.Export("sentry-project-slug", projectOutput.Slug)
		ctx.Export("sentry-project-default-client-key-dsn-public", projectOutput.DefaultClientKeyDSNPublic)
	}

	return nil
}

func getenvWithDefault(name, dflt string) string {
	if ret, ok := os.LookupEnv(name); ok {
		return ret
	}
	return dflt
}
