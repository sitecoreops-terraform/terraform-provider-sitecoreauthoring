# Sitecore Authoring Terraform provider

The project is a Terraform provider built using
* Go
* Terraform Plugin Framework

The folder structure:
* `pkg/apiclient/` which is a client to interact with the Sitecore authoring API, see [Query examples for authoring operastions](https://doc.sitecore.com/xp/en/developers/104/sitecore-experience-manager/query-examples-for-authoring-operations.html)
* `pkg/provider/` which is the Terraform provider that exposes resources and datasources and uses the apiclient to call the api.
* `examples/` with several terraform examples to show how the provider can be used in terraform modules.
* `docs/` contains the provider documentation for the Terraform registry. Everything here is automatically generated and should not be edited manually.

## Steps involved

When adding an endpoint / terraform resource you need to ensure or create that there are:

* Methods in apiclient for the endpoint
* Integration tests in apiclient that will call the actual endpoint based on environment variables
* Create unit test in apiclient working with mocked response
* Create resource in provider
* Create unit test in provider of resource schema
* Create example in `/examples/resources` folder to include in documentation
* If additional documentation is needed, create a template for the resource in `/templates/resources` folder, often this can be omitted. Do not edit files in `/docs` folder as those are generated.

## Initial setup

```bash
go install -v ./...
```

## Linting

```bash
golangci-lint run
```

## Formatting

```bash
gofmt -s -w -e .
```

## Building

```bash
go build
```


## Testing terraform provider

```sh
# Run a specific test, here client authentication
go test ./pkg/provider/... -v -run TestProviderMetadata

# Run all tests
go test ./pkg/provider/... -v
```

## Documentation

Documentation in `docs/` folder is for the public Terraform registry, see documentation at [Documenting terraform providers](https://developer.hashicorp.com/terraform/registry/providers/docs).

### Examples for resources and data sources

To create examples that will be included in the generated documentation pages, the examples should be created at the right paths. See [Examples readme](./examples/README.md)

### Generating Documentation

The documentation in `/docs` is generated using the `tfplugindocs` tool, see [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs), based on provider schema in the source files and the templates in `/templates` folder.

To generate documentation run:

```bash
cd tools
go generate
```

This will update the documentation files in the `docs/` directory based on the provider's Go code, the tempaltes, and the structured examples

## Using Sitecore CLI authentication

Sitecore CLI is a command line. The provider can re-use authentication from the CLI. To enabl

* [dotnet 8 SDK](https://dotnet.microsoft.com/en-us/download/dotnet/8.0) (as it is requirement for Sitecore CLI)
* [Sitecore CLI](https://doc.sitecore.com/xp/en/developers/latest/developer-tools/install-sitecore-command-line-interface.html) by running:
    ```bash
    dotnet nuget add source -n Sitecore https://nuget.sitecore.com/resources/v3/index.json
    dotnet tool install Sitecore.CLI
    ```
* Initialize Sitecore CLI with SitecoreAI plugin:
    ```bash
    dotnet sitecore init
    dotnet sitecore plugin add -n Sitecore.DevEx.Extensibility.XMCloud
    ```
* Authenticate by running
    ```bash
    dotnet sitecore cloud login
    ```
* Define to use Sitecore CLI authentication
    ```bash
    export SITECORE_AUTHORING_USE_CLI=1
    ```
