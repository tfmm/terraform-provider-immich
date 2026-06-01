# Terraform Provider Immich

A Terraform/OpenTofu provider for managing [Immich](https://immich.app/).

This provider and its documentation were developed with the assistance of Gemini, an AI assistant from Google.

Immich is a high-performance self-hosted photo and video management solution. This provider allows you to manage users, API keys, albums, and shared links programmatically.

## Documentation

Full documentation for the provider can be found on the [Terraform Registry](https://registry.terraform.io/providers/tfmm/immich/latest/docs).

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

## Installation

### Terraform

To use this provider in Terraform, add the following to your configuration:

```hcl
terraform {
  required_providers {
    immich = {
      source = "registry.terraform.io/tfmm/immich"
    }
  }
}

provider "immich" {
  # endpoint = "http://your-immich-instance:2283/api"
  # api_key  = "your-admin-api-key"
}
```

### OpenTofu

To use this provider in OpenTofu, add the following to your configuration:

```hcl
terraform {
  required_providers {
    immich = {
      source = "registry.opentofu.org/tfmm/immich"
    }
  }
}

provider "immich" {
  # endpoint = "http://your-immich-instance:2283/api"
  # api_key  = "your-admin-api-key"
}
```

The provider can be configured via environment variables:
- `IMMICH_ENDPOINT`: The full URL of the Immich API (e.g., `http://192.168.1.10:2283/api`)
- `IMMICH_API_KEY`: Your Immich API key.

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install .
```

## Documentation Generation

The documentation is generated using `terraform-plugin-docs`. To generate the documentation, run:

```shell
go generate ./...
```

(Note: This requires `tfplugindocs` to be installed and a `//go:generate` directive in `main.go`)

## License

MIT License - see the [LICENSE](LICENSE) file for details.
