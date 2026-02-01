package ccip

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// Regex pattern for parsing and matching a function signature.
	FunctionPattern = `function\s+(?<functionName>[a-zA-Z]{1}[a-zA-Z0-9\_]+)\s*\((?<input>[a-zA-Z0-9\s\,]*)\)\s+(?<mutability>pure|view)\s*returns\s+\((?<output>[a-zA-Z0-9\s\,]*)\)`
	FunctionRegex   = regexp.MustCompile(FunctionPattern)
)

var (
	ErrInvalidFunctionSignature = errors.New("invalid function signature")
)

type Variable struct {
	Name  string
	Type  string
	Value any
}

func (v Variable) Bytes32() (common.Hash, error) {
	switch val := v.Value.(type) {
	case common.Hash:
		return val, nil
	case [32]uint8:
		var hash common.Hash
		copy(hash[:], val[:])
		return hash, nil
	default:
		return common.Hash{}, fmt.Errorf("unsupported type for bytes32 conversion: %T", v.Value)
	}
}

// Takes a string like "function foo(uint256 bar) external view returns (uint256)"
// and parses it into a structured representation of the ABI.
func ParseFunction(str string) (*abi.Method, error) {
	if !FunctionRegex.MatchString(str) {
		return nil, ErrInvalidFunctionSignature
	}

	matches := FunctionRegex.FindStringSubmatch(str)

	functionName := matches[FunctionRegex.SubexpIndex("functionName")]
	mutability := matches[FunctionRegex.SubexpIndex("mutability")]
	input := matches[FunctionRegex.SubexpIndex("input")]
	output := matches[FunctionRegex.SubexpIndex("output")]

	parsedInputs, err := parseParameters(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input parameters: %w", err)
	}

	parsedOutputs, err := parseParameters(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output parameters: %w", err)
	}

	constructedMethod := abi.NewMethod(functionName, functionName, abi.Function, mutability, false, false, parsedInputs, parsedOutputs)
	return &constructedMethod, nil
}

func parseParameters(paramStr string) ([]abi.Argument, error) {
	// split at commas
	params := strings.Split(paramStr, ",")

	var paramNames = "abcdefghijklmnopqrstuvwxyz"

	var abiArgs []abi.Argument

	for i, param := range params {
		param = strings.TrimSpace(param)

		if param == "" {
			continue
		}

		fields := strings.Fields(param)

		if len(fields) < 1 || len(fields) > 2 {
			return nil, fmt.Errorf("invalid parameter format: %s", param)
		}

		paramType := fields[0]
		paramName := paramNames[i : i+1]

		if len(fields) == 2 {
			paramName = fields[1]
		}

		abiType, err := abi.NewType(paramType, paramType, nil)

		if err != nil {
			return nil, fmt.Errorf("invalid parameter type: %s", paramType)
		}

		abiArg := abi.Argument{
			Name:    paramName,
			Type:    abiType, // Type parsing can be added here if needed
			Indexed: false,
		}

		abiArgs = append(abiArgs, abiArg)
	}

	return abiArgs, nil
}
