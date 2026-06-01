# Contributing to Terraform Provider Immich

We love your contributions! Here's a quick guide on how to help out.

## Development Requirements

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

## Building

```shell
go build .
```

## Running Tests

To run the full suite of acceptance tests, you will need a running Immich instance.

```shell
# Set environment variables for the test instance
export IMMICH_ENDPOINT=http://localhost:2283/api
export IMMICH_API_KEY=your-admin-api-key

# Run acceptance tests
TF_ACC=1 go test ./... -v
```

## Documentation

Documentation is generated using [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs).

```shell
go generate ./...
```
