package commands

import "fmt"

func SerializeOutput(output any, isError bool) []byte {
	if output == "PONG" {
		return []byte(fmt.Sprintf("+%s\r\n", output))
	}

	if isError {
		return []byte(fmt.Sprintf("-%s\r\n", output))
	}

	switch v := output.(type) {
	case string:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int, int64, int32:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case []string:
		return serializeArrayOfStrings(v)

	case nil:
		return []byte("$-1\r\n")

	default:
		return nil
	}
}

func serializeString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func serializeArrayOfStrings(v []string) []byte {
	var result = fmt.Sprintf("*%d\r\n", len(v))
	for _, elem := range v {
		elemSerialized := serializeString(elem)
		result = result + elemSerialized
	}
	return []byte(result)

}
