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
	if os.Getenv("SKIP_PROJECT") == "" {
		_, err := sentry.NewProject(ctx, "testing", &sentry.ProjectArgs{
			Name:             pulumi.String("Sample Project"),
			Slug:             pulumi.String("sample-project-2"),
			OrganizationSlug: pulumi.String("ro-3w"),
			TeamSlug:         pulumi.String("ro"),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
