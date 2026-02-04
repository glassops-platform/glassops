IMAGE_NAME := glassops/salesforce-adapter
VERSION := v1.0.0

.PHONY: build run-info run-transform run-validate clean

build:
	docker build -t $(IMAGE_NAME):$(VERSION) .
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest

run-info:
	docker run --rm $(IMAGE_NAME):latest info

run-transform:
	docker run --rm -v $(PWD):/workspace $(IMAGE_NAME):latest \
		transform --input /workspace --output /workspace/output.sarif --policy-ref mock-policy

run-validate:
	docker run --rm -v $(PWD):/workspace $(IMAGE_NAME):latest \
		validate --input /workspace

clean:
	docker rmi $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest
