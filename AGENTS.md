# Terraform provider for Sitecore AUthoring API

- CONTRIBUTING.md

## Commands

### Linting
```bash
golangci-lint run
```

### Formatting
```bash
gofmt -s -w -e .
```

### Building
```bash
go build
```

### Testing
#### API Client
```bash
export SITECORE_AUTHORING_CLIENT_ID=your_client_id
export SITECORE_AUTHORING_CLIENT_SECRET=your_client_secret
export SITECORE_AUTHORING_HOST=your_authoring_instance_url
go test ./pkg/apiclient/... -v
```

#### Provider
```bash
go test ./pkg/provider/... -v
```

### Documentation Generation
```bash
cd tools && go generate
```

### Local Development Setup
```bash
go install -v ./...
```

### Debugging
```bash
export TF_LOG=DEBUG
```

### Authentication
#### Environment Variables
```bash
export SITECORE_AUTHORING_CLIENT_ID=your_client_id
export SITECORE_AUTHORING_CLIENT_SECRET=your_client_secret
```

#### Sitecore CLI
```bash
export SITECORE_AUTHORING_USE_CLI=1
```

