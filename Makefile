default:


test:
	rm -f profile.cov
	go test ./... -covermode=count -coverprofile=profile.cov;
	scripts/total_coverage.sh