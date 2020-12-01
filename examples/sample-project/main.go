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
	skipProject := os.Getenv("SKIP_PROJECT") != ""

	if !skipProject {
		_, err := sentry.NewProject(ctx, "testing", &sentry.ProjectArgs{
			Name:             pulumi.String(getenvWithDefault("PROJ_NAME", "Sample Project")),
			Slug:             pulumi.String(getenvWithDefault("PROJ_SLUG", "sample-project")),
			OrganizationSlug: pulumi.String(getenvWithDefault("ORG_SLUG", "ro-3w")),
			TeamSlug:         pulumi.String(getenvWithDefault("TEAM_SLUG", "ro")),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getenvWithDefault(name, dflt string) string {
	ret := os.Getenv(name)
	if ret != "" {
		return ret
	}
	return dflt
}
