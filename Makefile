GO ?= go
STATICCHECK ?= staticcheck
COVERPROFILE ?= coverage.out

.PHONY: ci coverage fmt fmt-check lint race test test-apply tidy-check verify vet

ci: fmt-check tidy-check verify vet test coverage race lint

fmt:
	gofmt -w .

fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "$$files"; \
		exit 1; \
	fi

tidy-check:
	$(GO) mod tidy -diff

verify:
	$(GO) mod verify

vet:
	$(GO) vet ./...

test:
	$(GO) test -count=1 ./...

test-apply:
	NONO_REQUIRE_APPLY=1 $(GO) test -run '^TestApplySandboxSubprocess$$' -count=1 -v ./...

coverage:
	NONO_SKIP_APPLY_SANDBOX_TEST=1 $(GO) test -count=1 -covermode=atomic -coverprofile=$(COVERPROFILE) ./...
	$(GO) tool cover -func=$(COVERPROFILE)

race:
	NONO_SKIP_APPLY_SANDBOX_TEST=1 $(GO) test -race -count=1 ./...

lint:
	$(STATICCHECK) ./...
