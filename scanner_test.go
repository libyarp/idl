package idl

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var file = `
package io.libyarp;

import "foo";
import "bar";

# This is a comment
# bound to RandomBytesRequest
message RandomBytesRequest {
    desired_length int8 = 0; # Fields indexes begin at zero.
}

message RandomBytesResponse {
    @repeated data uint8 = 0;
}

message MessageUsingExternalType {
	external_field io.libyarp.common.Foo = 0;
}

service RandomBytesService {
    generate_random_bytes(RandomBytesRequest) -> RandomBytesResponse;
}

service ServiceUsingExternalTypes {
	some_random_method(io.libyarp.common.External) -> RandomBytesResponse;
	some_random_method2(RandomBytesRequest) -> io.libyarp.common.External2;
	some_random_method3() -> io.libyarp.common.External2;
	some_method_without_response();
}
`

func TestParser(t *testing.T) {
	tokens, err := Scan(strings.NewReader(file))
	require.NoError(t, err)
	for i, t := range tokens {
		fmt.Printf("[%d] %s\n", i, t)
	}
	tree, err := Parse(tokens)
	require.NoError(t, err)

	assert.Equal(t, "io.libyarp", tree.Package)
	assert.Equal(t, []string{"RandomBytesRequest", "RandomBytesResponse", "MessageUsingExternalType"}, tree.DeclaredMessages)
	assert.Equal(t, []string{"RandomBytesService", "ServiceUsingExternalTypes"}, tree.DeclaredServices)
	assert.Equal(t, []string{"foo", "bar"}, tree.ImportedFiles)

	msg, ok := tree.MessageByName("RandomBytesRequest")
	assert.True(t, ok)
	assert.Equal(t, "RandomBytesRequest", msg.Name)
	assert.NotEmpty(t, msg.Comments)
}
