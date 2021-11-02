

build-mocks:
	cd pkg/platform/storage && mockery --all;
	cd pkg/staticfiles && mockery --all;