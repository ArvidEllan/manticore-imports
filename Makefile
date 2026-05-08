APP_NAME=manticore-imports
BINARY=bootstrap
PACKAGE=function.zip

.PHONY: build package clean deploy-dev deploy-prod

build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY) ./cmd/api

package: build
	zip -q -j $(PACKAGE) $(BINARY)

clean:
	rm -f $(BINARY) $(PACKAGE)

deploy-dev: package
	sls deploy --stage dev

deploy-prod: package
	sls deploy --stage prod
