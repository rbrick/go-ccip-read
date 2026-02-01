package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rbrick/go-ccip-read"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AddressRecord represents a domain name to address mapping in the database
type AddressRecord struct {
	gorm.Model
	Namehash string `gorm:"uniqueIndex;size:66"` // 0x prefixed 32 bytes hex
	Address  string `gorm:"size:42"`             // Ethereum address
	Owner    string `gorm:"size:42"`             // Owner of the record
}

// TextRecord represents text records for domains
type TextRecord struct {
	gorm.Model
	Namehash string `gorm:"index;size:66"`
	Key      string `gorm:"size:255"`
	Value    string `gorm:"type:text"`
}

func main() {
	// Initialize database
	db, err := gorm.Open(sqlite.Open("resolver.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&AddressRecord{}, &TextRecord{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed some example data
	seedDatabase(db)

	// Create CCIP Read resolver
	resolver := ccip.NewCCIPReadResolver()

	// Handle addr(bytes32) function - returns the address for a given namehash
	resolver.Handle("function addr(bytes32 namehash) view returns (address)", func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
		namehashVar, ok := request.Var("namehash")
		if !ok {
			return nil, fmt.Errorf("namehash parameter not found")
		}

		namehash := namehashVar.Value.([32]byte)
		namehashHex := "0x" + common.Bytes2Hex(namehash[:])

		var record AddressRecord
		result := db.Where("namehash = ?", namehashHex).First(&record)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Return zero address if not found
				return []interface{}{common.Address{}}, nil
			}
			return nil, fmt.Errorf("database error: %v", result.Error)
		}

		address := common.HexToAddress(record.Address)
		return []interface{}{address}, nil
	})

	// Handle text(bytes32, string) function - returns text records
	resolver.Handle("function text(bytes32 namehash, string key) view returns (string)", func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
		namehashVar, ok := request.Var("namehash")
		if !ok {
			return nil, fmt.Errorf("namehash parameter not found")
		}

		keyVar, ok := request.Var("key")
		if !ok {
			return nil, fmt.Errorf("key parameter not found")
		}

		namehash := namehashVar.Value.([32]byte)
		namehashHex := "0x" + common.Bytes2Hex(namehash[:])
		key := keyVar.Value.(string)

		var textRecord TextRecord
		result := db.Where("namehash = ? AND key = ?", namehashHex, key).First(&textRecord)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Return empty string if not found
				return []interface{}{""}, nil
			}
			return nil, fmt.Errorf("database error: %v", result.Error)
		}

		return []interface{}{textRecord.Value}, nil
	})

	// Start the server
	fmt.Println("CCIP Read resolver with GORM running on :8080")
	log.Fatal(http.ListenAndServe(":8080", resolver))
}

func seedDatabase(db *gorm.DB) {
	// Example namehashes (in a real application, these would be computed from actual domain names)
	exampleRecords := []AddressRecord{
		{
			Namehash: "0x0000000000000000000000000000000000000000000000000000000000000001",
			Address:  "0x1111111111111111111111111111111111111111",
			Owner:    "0xOwner1111111111111111111111111111111111",
		},
		{
			Namehash: "0x0000000000000000000000000000000000000000000000000000000000000002",
			Address:  "0x2222222222222222222222222222222222222222",
			Owner:    "0xOwner2222222222222222222222222222222222",
		},
	}

	exampleTextRecords := []TextRecord{
		{
			Namehash: "0x0000000000000000000000000000000000000000000000000000000000000001",
			Key:      "email",
			Value:    "user@example.com",
		},
		{
			Namehash: "0x0000000000000000000000000000000000000000000000000000000000000001",
			Key:      "url",
			Value:    "https://example.com",
		},
	}

	// Insert records if they don't exist
	for _, record := range exampleRecords {
		db.Where(AddressRecord{Namehash: record.Namehash}).FirstOrCreate(&record)
	}

	for _, record := range exampleTextRecords {
		db.Where(TextRecord{Namehash: record.Namehash, Key: record.Key}).FirstOrCreate(&record)
	}

	fmt.Println("Database seeded with example data")
}
