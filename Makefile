.PHONY: fmt vet test race web generated verify package

fmt:
	@test -z "$$(gofmt -l $$(git ls-files '*.go'))"

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

web:
	npm --prefix web/report test
	npm --prefix web/report run build

generated:
	./scripts/verify-generated.sh

verify: fmt vet test web generated

package:
	./scripts/package-release.sh
