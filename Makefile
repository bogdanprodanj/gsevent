.PHONY: vendor
# download project dependencies
vendor:
	GO111MODULE=on go mod vendor

# install all the necessary libraries for code generation
init: vendor
	GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger
	go get -u github.com/golang/mock/mockgen

# generate models
generate-swagger:
	swagger generate server --exclude-main --skip-operations --exclude-spec --skip-support --name=gsevent --target=./runtime

# generate mock files
generate-mocks:
	rm -rf mocks
	mkdir mocks
	$(GOPATH)/bin/mockgen --source=runtime/storage/storage.go --package=mocks --destination=mocks/mock_storage.go

generate: generate-swagger generate-mocks

run: init generate
	docker-compose up -d

