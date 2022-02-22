default:

test:
	rm -f profile.cov
	go test ./... -covermode=count -coverprofile=profile.cov;
	scripts/total_coverage.sh

server:
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY="true" \
	GIT_REPO_DRY_RUN="true" \
	GIT_REPO_BRANCH=main \
	LISTEN_ON="localhost:8081" \
	HEADER_SECRET="FAKE" \
	AZURE_SUBSCRIPTION_ID="e7220048-8a2c-4537-994b-6f9b320692d7" \
	go run main.go microservice server

docker-build-dev-platform-api:
	docker build -f ./Dockerfile -t dolittle-dev/platform-api:latest .

docker-build-dev-platform-operations: docker-build-dev-platform-api
	docker build --build-arg BASE_IMAGE=dolittle-dev/platform-api:latest -f Dockerfile.operations -t platform-operations:latest .

// TODO maybe change how we store this
build-tests: