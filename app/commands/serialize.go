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

func Encode(val any) (string, error) {
	switch v := val.(type) {
	case string:
		return BulkString(v), nil
	case int, int64, int32:
		return Number(v.(int)), nil
	case []string:
		return EncodeStringArray(v), nil
	case StreamValue:
		res := TypeArray + "2" + CRLF
		res += BulkString(v.Id)

		flatMap := make([]any, 0, len(v.Fields)*2)
		for key, v := range v.Fields {
			flatMap = append(flatMap, key)
			flatMap = append(flatMap, v)
		}

		encodedArr, err := Encode(flatMap)
		if err != nil {
			return "", err
		}
		res += encodedArr
		return res, nil
	case [][]any:
		res := TypeArray + fmt.Sprintf("%d", len(v)) + CRLF
		for _, item := range v {
			encoded, err := Encode(item)
			if err != nil {
				return "", err
			}
			res += encoded
		}
		return res, nil

	case []any:
		res := TypeArray + fmt.Sprintf("%d", len(v)) + CRLF
		for _, item := range v {
			encoded, err := Encode(item)
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

func EncodeStream(stream []StreamValue) string {
	return ""
}
