# CCIP Read Resolver Example with GORM

This example demonstrates how to build a CCIP Read resolver for ENS-like address resolution using GORM for database persistence.

## Features

- **Address Resolution**: Resolves Ethereum addresses from namehashes using the `addr(bytes32)` function
- **Text Records**: Supports text records like email and URL using the `text(bytes32, string)` function
- **SQLite Database**: Uses GORM with SQLite for persistent storage
- **Auto-seeding**: Automatically populates example data on startup

## Running the Example

1. Build the example:
```bash
go build
```

2. Run the server:
```bash
./examples
```

The server will start on port 8080 and create a `resolver.db` SQLite database file.

## Database Schema

### AddressRecord
- `namehash`: Unique 32-byte hash (hex-encoded with 0x prefix)
- `address`: Ethereum address (42 characters with 0x prefix)
- `owner`: Owner of the record

### TextRecord
- `namehash`: 32-byte hash (hex-encoded with 0x prefix)
- `key`: Text record key (e.g., "email", "url")
- `value`: Text record value

## Testing the Resolver

You can test the resolver by sending HTTP POST requests:

### Test addr() function:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "data": "0x3b3b57de0000000000000000000000000000000000000000000000000000000000000001"
  }'
```

This calls `addr(bytes32)` with namehash `0x00...01` and should return the address `0x1111111111111111111111111111111111111111`.

### Test text() function:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "data": "0x59d1d43c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000565616d61696c000000000000000000000000000000000000000000000000000000"
  }'
```

This calls `text(bytes32, string)` with namehash `0x00...01` and key "email" and should return "user@example.com".

## Extending the Example

You can extend this example by:
- Adding more ENS-like functions (contenthash, pubkey, etc.)
- Implementing access control for updating records
- Adding a web interface for managing records
- Supporting different database backends (PostgreSQL, MySQL)
- Implementing proper namehash computation from domain names
