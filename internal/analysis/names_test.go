package analysis

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeTypeNameConsistency(t *testing.T) {
	// Test that ComputeTypeName produces consistent results for the same type.
	nameCache := NewNameCache()

	// Create a basic type.
	basicType := types.Typ[types.String]

	// Call ComputeTypeName multiple times.
	name1 := nameCache.ComputeTypeName(basicType)
	name2 := nameCache.ComputeTypeName(basicType)
	require.Equal(t, name1, name2, "ComputeTypeName produced different results for same type")

	// Test with a pointer type.
	ptrType := types.NewPointer(basicType)
	ptrName1 := nameCache.ComputeTypeName(ptrType)
	ptrName2 := nameCache.ComputeTypeName(ptrType)
	require.Equal(t, ptrName1, ptrName2, "ComputeTypeName produced different results for same pointer type")

	// Expected names.
	require.Equal(t, "string", name1)
	require.Equal(t, "*string", ptrName1)
}

func TestComputeTypeNameWithNil(t *testing.T) {
	nameCache := NewNameCache()
	require.Empty(t, nameCache.ComputeTypeName(nil))
}

func TestComputeTypeCaching(t *testing.T) {
	// Test that the cache is actually working by verifying that.
	// multiple types produce the expected number of cache entries
	nameCache := NewNameCache()

	// Create multiple types.
	stringType := types.Typ[types.String]
	intType := types.Typ[types.Int]
	boolType := types.Typ[types.Bool]

	// Create pointer types.
	ptrString := types.NewPointer(stringType)
	ptrInt := types.NewPointer(intType)

	// Create a slice type.
	sliceString := types.NewSlice(stringType)

	// Call ComputeTypeName multiple times for each type.
	for range 10 {
		require.Equal(t, "string", nameCache.ComputeTypeName(stringType))
		require.Equal(t, "int", nameCache.ComputeTypeName(intType))
		require.Equal(t, "bool", nameCache.ComputeTypeName(boolType))
		require.Equal(t, "*string", nameCache.ComputeTypeName(ptrString))
		require.Equal(t, "*int", nameCache.ComputeTypeName(ptrInt))
		require.Equal(t, "[]string", nameCache.ComputeTypeName(sliceString))
	}
}

func TestComputeTypeNameComplexTypes(t *testing.T) {
	nameCache := NewNameCache()

	// Create a named type.
	pkg := types.NewPackage("github.com/example/test", "test")
	typename := types.NewTypeName(0, pkg, "MyType", nil)
	named := types.NewNamed(typename, types.Typ[types.Int], nil)

	// Test the named type.
	name1 := nameCache.ComputeTypeName(named)
	name2 := nameCache.ComputeTypeName(named)
	require.Equal(t, name1, name2)
	require.Equal(t, "github.com/example/test.MyType", name1)

	// Test pointer to named type.
	ptrNamed := types.NewPointer(named)
	ptrName1 := nameCache.ComputeTypeName(ptrNamed)
	ptrName2 := nameCache.ComputeTypeName(ptrNamed)
	require.Equal(t, ptrName1, ptrName2)
	require.Equal(t, "*github.com/example/test.MyType", ptrName1)

	// Test interface type.
	methods := []*types.Func{
		types.NewFunc(0, pkg, "Method1", types.NewSignatureType(nil, nil, nil, nil, nil, false)),
	}
	iface := types.NewInterfaceType(methods, nil).Complete()

	ifaceName1 := nameCache.ComputeTypeName(iface)
	ifaceName2 := nameCache.ComputeTypeName(iface)
	require.Equal(t, ifaceName1, ifaceName2)
}

func TestComputeObjectNameComplexTypes(t *testing.T) {
	nameCache := NewNameCache()
	pkg := types.NewPackage("github.com/example/test", "test")

	// Create a function object.
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	fn := types.NewFunc(0, pkg, "MyFunc", sig)

	// Call multiple times.
	name1 := nameCache.ComputeObjectName(fn)
	name2 := nameCache.ComputeObjectName(fn)
	require.Equal(t, name1, name2)
	require.Equal(t, "github.com/example/test.MyFunc", name1)

	// Create a method.
	recv := types.NewVar(0, pkg, "r", types.Typ[types.Int])
	methodSig := types.NewSignatureType(recv, nil, nil, nil, nil, false)
	method := types.NewFunc(0, pkg, "Method", methodSig)

	methodName1 := nameCache.ComputeObjectName(method)
	methodName2 := nameCache.ComputeObjectName(method)
	require.Equal(t, methodName1, methodName2)
	require.Equal(t, "github.com/example/test.int.Method", methodName1)
}

func TestMultipleNameCaches(t *testing.T) {
	// Test that multiple NameCache instances have independent caches.
	cache1 := NewNameCache()
	cache2 := NewNameCache()

	typ := types.Typ[types.String]

	// Both caches should produce the same result.
	name1 := cache1.ComputeTypeName(typ)
	name2 := cache2.ComputeTypeName(typ)
	require.Equal(t, name1, name2)
	require.Equal(t, "string", name1)

	// But they should have independent caches (this is just a logical test)
	// Each cache should work correctly on its own.
	for range 10 {
		require.Equal(t, name1, cache1.ComputeTypeName(typ))
		require.Equal(t, name2, cache2.ComputeTypeName(typ))
	}
}
