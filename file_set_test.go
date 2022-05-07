package idl

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFileSet(t *testing.T) {
	fs := NewFileSet()
	err := fs.Load("./test/fixture/test.yarp")
	require.NoError(t, err)
	fmt.Printf("%#v\n", fs)
	fmt.Printf("%#v\n", fs)
}
