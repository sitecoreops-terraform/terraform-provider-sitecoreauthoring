# Terraform provider for Sitecore AUthoring API

- CONTRIBUTING.md

## Testing Guidelines

When writing tests, follow these patterns:

1. **Individual Test Cases**: Each test case should have its own `t.Run()` function for better isolation and clearer test output.
2. **Avoid Table-Driven Loops**: Instead of using `for _, tc := range testCases` loops, restructure tests to use individual `t.Run()` calls.
3. **Clear Test Names**: Use descriptive names for each test case that clearly indicate what is being tested.

### Example Pattern

```go
func TestFunctionName(t *testing.T) {
    t.Run("Test case description", func(t *testing.T) {
        // Test setup
        input := "test"
        expected := "result"
        
        // Execute
        actual := functionUnderTest(input)
        
        // Verify
        if actual != expected {
            t.Errorf("Expected %s, got %s", expected, actual)
        }
    })
    
    t.Run("Another test case", func(t *testing.T) {
        // Another test case
    })
}
```

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

