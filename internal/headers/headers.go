package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var ERROR_INVALID_FIELD_VALUE = errors.New("space before colon")

const whitespace = " \t"

var (
	seperator = []byte("\r\n")
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Remove(key string) {
	delete(h, strings.ToLower(key))
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}

func (h Headers) Set(key, value string) {
	h[strings.ToLower(key)] = value
}

func parseHeaderLine(data []byte) (string, string, error) {
	var key, value string
	fieldName, fieldValue, ok := bytes.Cut(data, []byte{':'})
	if !ok {
		return key, value, nil
	}

	//Not allowed space before colon
	if bytes.ContainsAny(fieldName, whitespace) {
		return key, value, ERROR_INVALID_FIELD_VALUE
	}

	key = string(bytes.Trim(fieldName, whitespace))
	value = string(bytes.Trim(fieldValue, whitespace))

	return key, value, nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		lineIdx := bytes.Index(data[read:], seperator)

		if lineIdx == -1 {
			return read, false, nil
		}

		//End of headers as seperator at start of data
		if lineIdx == 0 {
			read += len(seperator)
			done = true
			break
		}

		key, value, err := parseHeaderLine(data[read : read+lineIdx])
		if err != nil {
			return 0, false, err
		}

		if valid := validFieldName([]byte(key)); !valid {
			return 0, false, fmt.Errorf("invalid character in field name")
		}
		lowerKey := strings.ToLower(key)
		if existing, ok := h[lowerKey]; ok {
			newVal := fmt.Sprintf("%s %s", existing, value)
			h[lowerKey] = newVal
		} else {
			h[lowerKey] = value

		}
		read += lineIdx + len(seperator)
	}

	return read, done, nil
}

var validSpecialChars = []byte("!#$%&'*+-.^_`|~")

func validFieldName(name []byte) bool {
	for _, b := range name {
		if b >= 'a' && b <= 'z' {
			continue
		}
		if b >= 'A' && b <= 'Z' {
			continue
		}
		if b >= '0' && b <= '9' {
			continue
		}
		if bytes.ContainsAny(validSpecialChars, string(b)) {
			continue
		}

		return false
	}

	return true
}
