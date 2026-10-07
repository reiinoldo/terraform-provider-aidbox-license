# Aidbox Provider

The Aidbox provider manages Aidbox licenses through the Aidbox RPC API. It uses
Terraform Plugin Protocol 6 and is intended to work with both Terraform and
OpenTofu.

## Requirements

- Terraform >= 1.0 or OpenTofu >= 1.6
- Go >= 1.23 to build from source

## Provider configuration

Set the API token either in provider configuration or with the
`AIDBOX_API_TOKEN` environment variable. Prefer the environment variable (or a
secret manager) instead of committing credentials to configuration.

The endpoint defaults to `https://aidbox.app/rpc`. Override it with the
`endpoint` provider argument or the `AIDBOX_ENDPOINT` environment variable.

```terraform
terraform {
  required_providers {
    aidbox = {
      source  = "reiinoldo/aidbox-license"
      version = "~> 0.1"
    }
  }
}

provider "aidbox" {}
```

```shell
export AIDBOX_API_TOKEN="..."
# Optional for non-default Aidbox installations:
export AIDBOX_ENDPOINT="https://aidbox.example.com/rpc"
```

## Managing a license

```terraform
resource "aidbox_license" "development" {
  name = "development"
  type = "development"
}
```

License name, product, and type changes replace the license. Destroying the
resource asks Aidbox to remove the license. The computed `jwt` attribute is
marked sensitive, but Terraform and OpenTofu still store it in state. Use an
encrypted, access-controlled remote state backend.

## Building and testing

```shell
go test ./...
go build .
```

The integration test in `internal/provider` creates a uniquely named license,
checks that it can be imported and read, then deletes it. It calls the live
Aidbox API and requires `AIDBOX_API_TOKEN`. Run it explicitly with:

```shell
$env:TF_ACC = "1"
$env:AIDBOX_API_TOKEN = "..."
go test -v -count=1 ./internal/provider -run '^TestAccAidboxLicenseResource$'
```

Set `AIDBOX_ENDPOINT` as well when testing a non-default Aidbox instance. The
same integration test can be run manually from the GitHub Actions workflow
after configuring the `AIDBOX_API_TOKEN` repository secret; the optional
`AIDBOX_ENDPOINT` repository variable overrides the default endpoint.

## Releasing

Create and push a version tag such as `v0.1.0` to trigger the GitHub release
workflow. Configure the `GPG_PRIVATE_KEY` and `PASSPHRASE` repository secrets
for signed release artifacts. Publishing the release in the Terraform Registry
additionally requires registering the provider repository under the `reiinoldo`
namespace.
