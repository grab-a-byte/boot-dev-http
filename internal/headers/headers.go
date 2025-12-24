package headers

import (
	"bytes"
	"errors"
)

var ERROR_INVALID_FIELD_VALUE = errors.New("space before colon")

const whitespace = " \t"

var (
	SEPERATOR = []byte("\r\n")
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func parseHeaderLine(data []byte) (string, string, error) {
	var key, value string
	colonIdx := bytes.IndexByte(data, ':')
	if colonIdx == -1 {
		return key, value, nil
	}
	fieldName := data[:colonIdx]
	fieldValue := data[colonIdx+1:]

	//Not allowed space before colon
	if bytes.ContainsAny(fieldName, whitespace) {
		return key, value, ERROR_INVALID_FIELD_VALUE
	}

	key = string(bytes.Trim(fieldName, whitespace))
	value = string(bytes.Trim(fieldValue, whitespace))

	return key, value, nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	read := 0
	for {
		lineIdx := bytes.Index(data[read:], SEPERATOR)

		if lineIdx == -1 {
			return 0, false, nil
		}

		//End of headers as seperator at start of data
		if lineIdx == 0 {
			return read + len(SEPERATOR), true, nil
		}

		key, value, err := parseHeaderLine(data[read : read+lineIdx])
		if err != nil {
			return 0, false, err
		}
		h[key] = value
		read += lineIdx + len(SEPERATOR)
	}
}
