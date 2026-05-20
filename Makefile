APP_NAME=manticore-imports
BINARY=bootstrap
PACKAGE=function.zip
STAGE?=dev

.PHONY: build package clean deploy-dev deploy-prod deploy-frontend-dev deploy-frontend-prod test

build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY) ./cmd/api

package: build
	zip -q -j $(PACKAGE) $(BINARY)

clean:
	rm -f $(BINARY) $(PACKAGE)

test:
	go test ./...

deploy-dev: package
	sls deploy --stage dev

deploy-prod: package
	sls deploy --stage prod

deploy-frontend-dev:
	bash scripts/deploy-frontend.sh dev

deploy-frontend-prod:
	bash scripts/deploy-frontend.sh prod

deploy-all-dev: deploy-dev deploy-frontend-dev

deploy-all-prod: deploy-prod deploy-frontend-prod
