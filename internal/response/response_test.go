package response_test

import (
	"strings"
	"testing"

	"dev.grab-a-byte.network/internal/response"
)

func TestResponseWriteHeaderLine(t *testing.T) {
	builder := strings.Builder{}
	response.WriteStatusLine(&builder, response.STATUS_OK)
	if builder.String() != "HTTP/1.1 200 OK\r\n" {
		t.Error(builder.String())
	}
}
