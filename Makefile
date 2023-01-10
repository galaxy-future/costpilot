prepare:
	go install golang.org/x/tools/cmd/goimports@latest

format:
	go fmt ./...
	find . -name '*.go' | grep -Ev 'vendor|thrift_gen' | xargs goimports -w

vet:
	@echo "go vet ."
	@go vet $$(go list ./...) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

check: vet format

clean:
	rm -rf output

build:
	sh scripts/build.sh

run:
	sh ./output/run.sh

#USE make TARGET version=xx override version
version ?= latest
docker-build:
	docker build -t costpilot:latest -f Dockerfile ./

docker-tag:
	docker tag costpilot:latest galaxyfuture/costpilot:${version}

docker-push-hub:
	docker push galaxyfuture/costpilot:${version}

docker-hub-all: docker-build docker-tag docker-push-hub