IMAGE := "kuberhealthy/resource-quota-check"
TAG := "latest"

# Build the resource quota check container locally.
build:
	podman build -f Containerfile -t {{IMAGE}}:{{TAG}} .

# Run the unit tests for the resource quota check.
test:
	go test ./...

# Build the resource quota check binary locally.
binary:
	go build -o bin/resource-quota-check ./cmd/resource-quota-check
