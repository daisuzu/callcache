.PHONY: test
test:
	@go test -race ./...

.PHONY: go-oldstable
go-oldstable:
	$(eval OLDSTABLE=$(shell curl -s 'https://go.dev/dl/?mode=json' | \
		grep -oE '"version": "go[0-9]+\.[0-9]+' | sed 's/"version": "go//' | sort -V | uniq | tail -n 2 | head -n 1))
	@if [ -z "$(OLDSTABLE)" ]; then \
		echo "Error: Could not determine oldstable version from JSON."; exit 1; \
	fi
	go get go@$(OLDSTABLE) && go mod tidy
