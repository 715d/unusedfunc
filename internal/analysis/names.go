package analysis

import (
	"go/types"
	"strings"

	"github.com/puzpuzpuz/xsync/v4"
)

// NameCache provides efficient caching of fully-qualified symbol names
// for deduplication across generic instantiations and type analysis.
type NameCache struct {
	objCache  *xsync.Map[types.Object, string]
	typeCache *xsync.Map[types.Type, string]
}

func NewNameCache() *NameCache {
	return &NameCache{
		objCache:  xsync.NewMap[types.Object, string](),
		typeCache: xsync.NewMap[types.Type, string](),
	}
}

// ComputeObjectName generates a canonical name for an Object.
// For functions, returns packagePath.objName() or packagePath.objName[T] for generics.
// For methods, includes receiver type information (e.g., "packagePath.Container[T].Clear" or "packagePath.Person.GetName").
func (c *NameCache) ComputeObjectName(obj types.Object) string {
	if obj == nil {
		return ""
	}
	name, ok := c.objCache.Load(obj)
	if ok {
		return name
	}
	name = c.computeObjectName(obj)
	c.objCache.Store(obj, name)
	return name
}

// ComputeTypeName generates a canonical name for a types.Type.
// For named types, returns packagePath.TypeName[TypeParams] (if generic)
// For pointer types, returns packagePath.*TypeName.
// For other types, returns the string representation with package path when available.
func (c *NameCache) ComputeTypeName(typ types.Type) string {
	if typ == nil {
		return ""
	}
	name, ok := c.typeCache.Load(typ)
	if ok {
		return name
	}
	name = c.computeTypeName(typ)
	c.typeCache.Store(typ, name)
	return name
}

func (c *NameCache) computeObjectName(obj types.Object) string {
	baseName := obj.Name()

	// Get package path.
	var packagePath string
	if pkg := obj.Pkg(); pkg != nil {
		packagePath = pkg.Path()
	}

	// Pre-allocate builder with estimated capacity.
	var builder strings.Builder
	builder.Grow(128) // Pre-allocate for typical object names

	// Add package path if present.
	if packagePath != "" {
		builder.WriteString(packagePath)
		builder.WriteByte('.')
	}

	// For methods, include receiver type information.
	if fn, ok := obj.(*types.Func); ok {
		fnType := fn.Type()
		if fnType == nil {
			builder.WriteString(baseName)
			return builder.String()
		}
		sig, ok := fnType.(*types.Signature)
		if !ok {
			builder.WriteString(baseName)
			return builder.String()
		}
		if recv := sig.Recv(); recv != nil {
			// Extract receiver type name.
			recvType := recv.Type()
			isPointer := false
			// Check if pointer receiver.
			if ptr, ok := recvType.(*types.Pointer); ok {
				recvType = ptr.Elem()
				isPointer = true
			}

			// Get the receiver type name.
			recvTypeName := getGenericTypeName(recvType)

			// Extract just the type name part (remove package path)
			if strings.Contains(recvTypeName, ".") {
				typeParts := strings.Split(recvTypeName, ".")
				recvTypeName = typeParts[len(typeParts)-1]
			}

			// Add pointer indicator if receiver is a pointer.
			if isPointer {
				builder.WriteByte('*')
			}
			builder.WriteString(recvTypeName)
			builder.WriteByte('.')
			builder.WriteString(baseName)
			return builder.String()
		}

		// For generic functions (not methods), check if signature has type parameters.
		if sig.TypeParams() != nil && sig.TypeParams().Len() > 0 {
			// Build the generic function name with type parameters.
			builder.WriteString(baseName)
			formatTypeParamsToBuilder(&builder, sig.TypeParams())
			return builder.String()
		}
	}

	// For regular functions and other objects.
	builder.WriteString(baseName)
	return builder.String()
}

func (c *NameCache) computeTypeName(typ types.Type) string {
	// Handle pointer types.
	if ptr, ok := typ.(*types.Pointer); ok {
		elemName := c.ComputeTypeName(ptr.Elem())
		if elemName == "" {
			return ""
		}
		var builder strings.Builder
		builder.Grow(len(elemName) + 1)
		builder.WriteByte('*')
		builder.WriteString(elemName)
		return builder.String()
	}

	// Handle named types (structs, interfaces, etc.)
	if named, ok := typ.(*types.Named); ok {
		obj := named.Obj()
		if obj == nil {
			return typ.String()
		}

		// Pre-allocate builder.
		var builder strings.Builder
		builder.Grow(128) // Pre-allocate for typical type names

		// Add package path.
		if pkg := obj.Pkg(); pkg != nil {
			builder.WriteString(pkg.Path())
			builder.WriteByte('.')
		}

		typeName := getGenericTypeName(typ)
		builder.WriteString(typeName)
		result := builder.String()

		return result
	}

	// For other types (basic types, slices, maps, etc.), return string representation.
	// These typically don't have package information.
	return typ.String()
}

// getGenericTypeName extracts the type name, preserving generic template syntax
// For generic types, returns the template form (e.g., "Container[T]")
// For non-generic types, returns the type string as-is.
func getGenericTypeName(typ types.Type) string {
	// Check if this is a named type.
	if named, ok := typ.(*types.Named); ok {
		// Get the type name.
		typeName := named.Obj().Name()

		// Check if this is an instantiated generic (has type arguments) FIRST.
		// This takes precedence over type parameters.
		if named.TypeArgs() != nil && named.TypeArgs().Len() > 0 {
			// Build name with actual type arguments.
			result := typeName + formatTypeArgs(named.TypeArgs())
			return result
		}

		// Check if this is a generic type template (has type parameters but no arguments)
		if named.TypeParams() != nil && named.TypeParams().Len() > 0 {
			return typeName + formatTypeParams(named.TypeParams())
		}

		// For non-generic types, just return the name.
		return typeName
	}

	// For other types, return the string representation.
	return typ.String()
}

// formatTypeParams formats a type parameter list into "[T, U, ...]" format
func formatTypeParams(typeParams *types.TypeParamList) string {
	if typeParams == nil || typeParams.Len() == 0 {
		return ""
	}
	var builder strings.Builder
	builder.Grow(32) // Pre-allocate for typical type params
	formatTypeParamsToBuilder(&builder, typeParams)
	return builder.String()
}

// formatTypeParamsToBuilder writes type parameters to an existing builder
func formatTypeParamsToBuilder(builder *strings.Builder, typeParams *types.TypeParamList) {
	if typeParams == nil || typeParams.Len() == 0 {
		return
	}
	builder.WriteByte('[')
	for i := range typeParams.Len() {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(typeParams.At(i).Obj().Name())
	}
	builder.WriteByte(']')
}

// formatTypeArgs formats a type argument list into "[T, string, ...]" format
func formatTypeArgs(typeArgs *types.TypeList) string {
	if typeArgs == nil || typeArgs.Len() == 0 {
		return ""
	}
	var builder strings.Builder
	builder.Grow(32) // Pre-allocate for typical type args
	builder.WriteByte('[')
	for i := range typeArgs.Len() {
		if i > 0 {
			builder.WriteString(", ")
		}
		arg := typeArgs.At(i)
		builder.WriteString(arg.String())
	}
	builder.WriteByte(']')
	return builder.String()
}
