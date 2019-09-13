DATASOURCE=csv-datasource
GO = GO111MODULE=on go

vendor: ## Vendor Go dependencies
	$(GO) mod vendor

build-linux: ## Build backend plugin for Linux
	$(GO) build -mod=vendor -o ./dist/${DATASOURCE}_linux_amd64 ./pkg

build-darwin: ## Build backend plugin for macOS
	$(GO) build -mod=vendor -o ./dist/${DATASOURCE}_darwin_amd64 ./pkg

build-win: ## Build backend plugin for Windows
	$(GO) build -mod=vendor -o ./dist/${DATASOURCE}_windows_amd64.exe ./pkg

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL=help
.PHONY=help
