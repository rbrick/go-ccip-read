# go-ccip-read

A simple library for creating CCIP-Read (EIP-3668) gateways easily in Go.

## Overview

`go-ccip-read` provides a clean, type-safe API for building CCIP-Read gateways in Go. It handles all the complexity of ABI encoding/decoding, HTTP request handling, and function signature parsing, allowing you to focus on implementing your business logic.

## Features

- 🚀 Simple, intuitive API for defining CCIP-Read handlers
- 📦 Automatic ABI encoding/decoding
- 🎯 Type-safe function signature parsing
- 🔧 Flexible configuration via functional options
- ⚡ Works as a standard `http.Handler` - integrate with any Go HTTP server

## Installation

```bash
go get github.com/rbrick/go-ccip-read
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    
    "github.com/ethereum/go-ethereum/common"
    ccip "github.com/rbrick/go-ccip-read"
)

func main() {
    // Create a new CCIP Read resolver
    resolver := ccip.NewCCIPReadResolver()
    
    // Register a handler for the addr(bytes32) function
    resolver.Handle("function addr(bytes32 namehash) view returns (address)", 
        func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
            // Get the namehash parameter
            namehashVar, ok := request.Var("namehash")
            if !ok {
                return nil, fmt.Errorf("namehash parameter not found")
            }
            
            // Your business logic here
            namehash := namehashVar.Value.([32]byte)
            address := lookupAddress(namehash)
            
            return []interface{}{address}, nil
        })
    
    // Start the server
    log.Fatal(http.ListenAndServe(":8080", resolver))
}
```

## Usage

### Creating a Resolver

Create a new CCIP-Read resolver with optional configuration:

```go
resolver := ccip.NewCCIPReadResolver(
    ccip.GatewayValidator(myValidator),
    // Add more options as needed
)
```

### Registering Handlers

Register handlers for Solidity function signatures. The library automatically handles ABI encoding/decoding:

```go
// Handle a simple address lookup
resolver.Handle("function addr(bytes32 namehash) view returns (address)", 
    func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
        namehashVar, _ := request.Var("namehash")
        namehash := namehashVar.Value.([32]byte)
        
        // Your lookup logic
        address := common.HexToAddress("0x1234...")
        
        return []interface{}{address}, nil
    })

// Handle a text record lookup with multiple parameters
resolver.Handle("function text(bytes32 namehash, string key) view returns (string)", 
    func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
        namehashVar, _ := request.Var("namehash")
        keyVar, _ := request.Var("key")
        
        namehash := namehashVar.Value.([32]byte)
        key := keyVar.Value.(string)
        
        // Your lookup logic
        value := lookupText(namehash, key)
        
        return []interface{}{value}, nil
    })
```

### Accessing Request Parameters

Access function parameters using the `Var` method:

```go
func handler(request *ccip.CCIPReadRequest) ([]interface{}, error) {
    // Get a parameter by name
    namehashVar, ok := request.Var("namehash")
    if !ok {
        return nil, fmt.Errorf("parameter not found")
    }
    
    // Access the value with appropriate type assertion
    namehash := namehashVar.Value.([32]byte)
    
    // Work with the parameter
    // ...
}
```

### Configuration Options

#### Gateway Validation

Restrict which addresses can call your gateway:

```go
// Allow specific gateway addresses
resolver := ccip.NewCCIPReadResolver(
    ccip.Gateways(
        common.HexToAddress("0x1111111111111111111111111111111111111111"),
        common.HexToAddress("0x2222222222222222222222222222222222222222"),
    ),
)

// Or provide a custom validator
resolver := ccip.NewCCIPReadResolver(
    ccip.GatewayValidator(func(sender common.Address) error {
        if !isAuthorized(sender) {
            return fmt.Errorf("unauthorized sender")
        }
        return nil
    }),
)
```

#### Custom Output Encoding

Provide a custom output encoder if needed:

```go
resolver := ccip.NewCCIPReadResolver(
    ccip.OutputEncoding(func(outputs []interface{}) ([]byte, error) {
        // Custom encoding logic
        return encode(outputs), nil
    }),
)
```

## Complete Example

See the [examples directory](examples/) for a complete example that demonstrates:

- Setting up a CCIP-Read gateway with GORM/SQLite
- Handling multiple function signatures
- Database integration
- Parameter extraction and type handling

Run the example:

```bash
cd examples
go run main.go
```

## HTTP API

The resolver exposes a single POST endpoint that accepts CCIP-Read requests:

**Request:**
```json
{
  "data": "0x...",      // ABI-encoded function call
  "sender": "0x..."     // (optional) Address of the calling contract
}
```

**Response:**
```json
{
  "data": "0x..."       // ABI-encoded return values
}
```

## Function Signature Format

Function signatures must follow this format:

```
function <name>(<param_type> <param_name>, ...) view returns (<return_type>, ...)
```

Examples:
- `function addr(bytes32 namehash) view returns (address)`
- `function text(bytes32 namehash, string key) view returns (string)`
- `function contenthash(bytes32 node) view returns (bytes)`

## Type Handling

Common Ethereum types are automatically handled:

| Solidity Type | Go Type |
|--------------|---------|
| `address` | `common.Address` |
| `bytes32` | `[32]byte` |
| `string` | `string` |
| `bytes` | `[]byte` |
| `uint256` | `*big.Int` |
| `bool` | `bool` |

## Error Handling

Handlers should return errors when something goes wrong:

```go
resolver.Handle("function addr(bytes32 namehash) view returns (address)", 
    func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
        namehashVar, ok := request.Var("namehash")
        if !ok {
            return nil, fmt.Errorf("namehash parameter not found")
        }
        
        result, err := lookupAddress(namehashVar.Value.([32]byte))
        if err != nil {
            return nil, fmt.Errorf("lookup failed: %w", err)
        }
        
        return []interface{}{result}, nil
    })
```

Errors are automatically returned to the client with an appropriate HTTP status code.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)

## Resources

- [EIP-3668: CCIP Read](https://eips.ethereum.org/EIPS/eip-3668)
- [ENS CCIP-Read Documentation](https://docs.ens.domains/resolvers/ccip-read)