help: ## Display this help
	@ echo "Please use \`make <target>' where <target> is one of:"
	@ echo
	@ grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-10s\033[0m - %s\n", $$1, $$2}'
	@ echo

outdated: ## Check outdated deps
	go list -u -m -mod=mod -json all | go-mod-outdated -update -direct

lint: ## Lint all the code
	golangci-lint run --timeout 5m

fix-lint: ## Fix the lint issues in the code (if possible)
	golangci-lint run --timeout 5m --fix
