package idl

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func assertField(t *testing.T, fv interface{}, assertions ...func(t *testing.T, f Field)) {
	assert.IsType(t, Field{}, fv)
	f := fv.(Field)
	for _, fn := range assertions {
		fn(t, f)
	}
}

func name(n string) func(t *testing.T, f Field) {
	return func(t *testing.T, f Field) {
		assert.Equal(t, f.Name, n)
	}
}

func optional() func(t *testing.T, f Field) {
	return func(t *testing.T, f Field) {
		_, ok := f.Annotations.FindByName(OptionalAnnotation)
		assert.True(t, ok, "expected field to have optional annotation")
	}
}

func repeat() func(t *testing.T, f Field) {
	return func(t *testing.T, f Field) {
		_, ok := f.Annotations.FindByName(RepeatedAnnotation)
		assert.True(t, ok, "expected field to have repeated annotation")
	}
}

func tInt8() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Int8, f.Type.(Primitive).Kind)
	}
}
func tInt16() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Int16, f.Type.(Primitive).Kind)
	}
}
func tInt32() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Int32, f.Type.(Primitive).Kind)
	}
}
func tInt64() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Int64, f.Type.(Primitive).Kind)
	}
}
func tUint8() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Uint8, f.Type.(Primitive).Kind)
	}
}
func tUint16() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Uint16, f.Type.(Primitive).Kind)
	}
}
func tUint32() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Uint32, f.Type.(Primitive).Kind)
	}
}
func tUint64() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, Uint64, f.Type.(Primitive).Kind)
	}
}
func tString() func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		require.IsType(t, Primitive{}, f.Type)
		assert.Equal(t, String, f.Type.(Primitive).Kind)
	}
}
func tStruct(typ string) func(*testing.T, Field) {
	return func(t *testing.T, f Field) {
		assert.Equal(t, TypeUnresolved, f.Type.Type())
		assert.Equal(t, typ, f.Type.(Unresolved).Name)
	}
}

func assertMethod(t *testing.T, m Method, assertions ...func(t *testing.T, m Method)) {
	for _, fn := range assertions {
		fn(t, m)
	}
}

func methodName(name string) func(t *testing.T, m Method) {
	return func(t *testing.T, m Method) {
		assert.Equal(t, name, m.Name)
	}
}

func streams() func(t *testing.T, m Method) {
	return func(t *testing.T, m Method) {
		assert.True(t, m.ReturnStreaming, "expected method to stream response")
	}
}

func argumentType(name string) func(t *testing.T, m Method) {
	return func(t *testing.T, m Method) {
		assert.Equal(t, name, m.ArgumentType)
	}
}

func returnType(name string) func(t *testing.T, m Method) {
	return func(t *testing.T, m Method) {
		assert.Equal(t, name, m.ReturnType)
	}
}

func TestParserDocsExample(t *testing.T) {
	f, err := os.Open("./test/fixture/contacts.yarp")
	require.NoError(t, err)
	tokens, err := Scan(f)
	require.NoError(t, f.Close())
	require.NoError(t, err)
	ff, err := Parse(tokens)
	require.NoError(t, err)

	joinComments := func(str []string) string {
		return strings.Join(str, " ")
	}

	t.Run("package", func(t *testing.T) {
		v := ff.Tree[0]
		require.IsType(t, Package{}, v)
		vv := v.(Package)
		assert.Equal(t, "org.example.contacts", vv.Name)
	})

	t.Run("Contact", func(t *testing.T) {
		v := ff.Tree[1]
		require.IsType(t, Message{}, v)
		vv := v.(Message)
		assert.Equal(t, "Contact", vv.Name)
		assert.Equal(t, "Contact represent a single person in the address list.", joinComments(vv.Comments))
		assert.Empty(t, vv.Annotations)
		assertField(t, vv.Fields[0], name("id"), optional(), tInt64())
		assertField(t, vv.Fields[1], name("name"), tString())
		assertField(t, vv.Fields[2], name("surname"), tString())
		assertField(t, vv.Fields[3], name("company"), optional(), tStruct("Company"))
		assertField(t, vv.Fields[4], name("emails"), repeat(), tString())
	})

	t.Run("Company", func(t *testing.T) {
		v := ff.Tree[2]
		require.IsType(t, Message{}, v)
		vv := v.(Message)
		assert.Equal(t, "Company", vv.Name)
		assert.Equal(t, "Company represents a company in which a person works at.", joinComments(vv.Comments))
		assertField(t, vv.Fields[0], name("name"), tString())
		assertField(t, vv.Fields[1], name("website_address"), tString())
	})

	t.Run("GetContactRequest", func(t *testing.T) {
		v := ff.Tree[3]
		require.IsType(t, Message{}, v)
		vv := v.(Message)
		assert.Equal(t, "GetContactRequest", vv.Name)
		assert.Equal(t, "GetContactRequest represents a request to obtain a specific contact through a given id.", joinComments(vv.Comments))
		assertField(t, vv.Fields[0], name("id"), tInt64())
	})

	t.Run("GetContactResponse", func(t *testing.T) {
		v := ff.Tree[4]
		require.IsType(t, Message{}, v)
		vv := v.(Message)
		assert.Equal(t, "GetContactResponse", vv.Name)
		assert.Equal(t, "GetContactResposne represents the result of a GetContactRequest. An absent `contact` indicates that no contact under the provided id exists.", joinComments(vv.Comments))
		assertField(t, vv.Fields[0], name("contact"), tStruct("Contact"), optional())
	})

	t.Run("ContactsService", func(t *testing.T) {
		v := ff.Tree[5]
		require.IsType(t, Service{}, v)
		vv := v.(Service)
		assert.Equal(t, "ContactsService", vv.Name)
		assertMethod(t, vv.Methods[0], methodName("upsert_contact"), argumentType("Contact"), returnType("void"))
		assertMethod(t, vv.Methods[1], methodName("list_contacts"), argumentType("void"), returnType("Contact"), streams())
		assertMethod(t, vv.Methods[2], methodName("get_contact"), argumentType("GetContactRequest"), returnType("GetContactResponse"))
	})
}
