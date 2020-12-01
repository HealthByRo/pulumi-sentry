all:
	echo TODO

.PHONY: go-install-provider
go-install-provider:
	go install ./cmd/pulumi-resource-sentry

rebuild-sdk:
	go install ./cmd/pulumi-sdkgen-sentry
	rm -rf ./sdk && pulumi-sdkgen-sentry ./schema.json ./sdk

.PHONY: sample-preview
sample-preview: go-install-provider rebuild-sdk
	pulumi -C examples/sample-project/ preview

.PHONY: sample-up
sample-up: go-install-provider rebuild-sdk
	pulumi -C examples/sample-project/ up
