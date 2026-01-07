package commands

import (
	"fmt"
)

const (
	// Protocol Delimiter
	CRLF = "\r\n"

	// Data Types (RESP2 & RESP3)
	TypeSimpleString = "+"
	TypeSimpleError  = "-"
	TypeInteger      = ":"
	TypeBulkString   = "$"
	TypeArray        = "*"

	// RESP3 Specific (Your BULK_ERROR "!" is a RESP3 Blob Error)
	TypeBlobError = "!"
	TypeVerbatim  = "="
	TypeMap       = "%"
	TypeSet       = "~"

	// Common Responses
	OK     = "+OK\r\n"
	QUEUED = "+QUEUED\r\n"

	// Nulls
	NullBulkString = "$-1\r\n"
	NullArray      = "*-1\r\n"
)

// func SerializeOutput(commandName string, output any, isError bool) []byte {
// 	if commandName == "PING" || commandName == "TYPE" || commandName == "SET" {
// 		return []byte(fmt.Sprintf("+%s\r\n", output))
// 	}

// 	if isError {
// 		return []byte(fmt.Sprintf("-%s\r\n", output))
// 	}

// 	switch v := output.(type) {
// 	case string:
// 		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
// 	case int, int64, int32:
// 		return []byte(fmt.Sprintf(":%d\r\n", v))
// 	case []string:
// 		return serializeArrayOfStrings(v)

// 	case nil:
// 		return []byte("$-1\r\n")

// 	default:
// 		return nil
// 	}
// }

// func serializeString(s string) string {
// 	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
// }

// func serializeArrayOfStrings(v []string) []byte {
// 	if len(v) == 1 && v[0] == "-1" {
// 		return []byte(fmt.Sprintf("*%d\r\n", -1))
// 	}
// 	var result = fmt.Sprintf("*%d\r\n", len(v))
// 	for _, elem := range v {
// 		elemSerialized := serializeString(elem)
// 		result = result + elemSerialized
// 	}
// 	return []byte(result)

// }

func Encode(val any) (string, error) {
	switch v := val.(type) {
	case string:
		return BulkString(v), nil
	case int, int64, int32:
		return Number(v.(int)), nil
	case []string:
		return EncodeStringArray(v), nil
	case []any:
		res := TypeArray + fmt.Sprintf("%d", len(v)) + CRLF
		for _, item := range v {
			encoded, err := Encode(item) // Recursive call
			if err != nil {
				return "", err
			}
			res += encoded
		}
		return res, nil

	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}

}

func EncodeError(error string) string {
	return fmt.Sprintf("%s%s%s", TypeSimpleError, error, CRLF)
}

func SimpleString(val string) string {
	return TypeSimpleString + val + CRLF
}

func BulkString(val string) string {
	return fmt.Sprintf("%s%d%s%s%s", TypeBulkString, len(val), CRLF, val, CRLF)
}

func Number(val int) string {
	return fmt.Sprintf("%s%d%s", TypeInteger, val, CRLF)
}

func EncodeStringArray(arr []string) string {
	// if len(arr) == 1 && arr[0] == "-1" {
	// 	return NullBulkString
	// }

	var result = fmt.Sprintf("*%d", len(arr)) + CRLF
	for _, elem := range arr {
		elemSerialized := BulkString(elem)
		result = result + elemSerialized
	}
	return result
}
