all:
	echo TODO

.PHONY: go-install-provider
go-install-provider:
	go install ./cmd/pulumi-resource-sentry
	go install ./cmd/pulumi-sdkgen-sentry

rebuild-sdk: go-install-provider
	rm -rf ./sdk && pulumi-sdkgen-sentry ./schema.json ./sdk
