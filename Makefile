.PHONY: bootstrap
bootstrap:
	script/install-glide
	glide install

.PHONY: update
update:
	glide up
	glide install

.PHONY: tests
tests:
	LOG_LEVEL=debug go test -cover -race $(shell glide nv)

.PHONY: docker-tests
docker-tests:
	docker run goruntime make tests
