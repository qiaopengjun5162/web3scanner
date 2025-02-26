// Package serializers provides a GORM serializer for the `[]byte` type.
package serializers

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type BytesSerializer struct{}

// BytesInterface is an interface that requires implementing a method Bytes() which returns a byte slice.
// It can be used to represent objects that can be converted to a byte slice representation.
type BytesInterface interface{ Bytes() []byte }

// SetBytesInterface is an interface that requires implementing a method SetBytes()
// which sets the byte slice value. It can be used to represent objects that
// can have their byte slice value modified.
type SetBytesInterface interface{ SetBytes([]byte) }

func init() {
	schema.RegisterSerializer("bytes", BytesSerializer{})
}

// Scan deserializes a database value into a field of type `[]byte` or a type that implements
// the `SetBytes([]byte)` interface.
//
// If the database value is nil, it will return nil.
//
// It first checks if the database value is a string. If not, it will return an error.
//
// If the database value is a string, it will attempt to decode it as a hex string using
// `hexutil.Decode`. If the decoding fails, it will return an error.
//
// If the decoding is successful, it will create a new value of the field type and call
// `SetBytes` on it with the decoded byte slice. If the field type does not implement the
// `SetBytes([]byte)` interface, it will return an error.
//
// If the field type is a pointer, it will detect it and allocate memory to where the
// allocated pointer should point to. If the field type is a double pointer, it will return
// an error.
//
// Finally, it will set the deserialized value into the dst value using `ReflectValueOf`.
func (BytesSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}

	hexStr, ok := dbValue.(string)
	if !ok {
		return fmt.Errorf("expected hex string as the database value: %T", dbValue)
	}

	b, err := hexutil.Decode(hexStr)
	if err != nil {
		return fmt.Errorf("failed to decode database value: %w", err)
	}

	fieldValue := reflect.New(field.FieldType)
	fieldInterface := fieldValue.Interface()

	// Detect if we're deserializing into a pointer. If so, we'll need to
	// also allocate memory to where the allocated pointer should point to
	if field.FieldType.Kind() == reflect.Pointer {
		nestedField := fieldValue.Elem()
		if nestedField.Elem().Kind() == reflect.Pointer {
			return fmt.Errorf("double pointers are the max depth supported: %T", fieldValue)
		}

		// We'll want to call `SetBytes` on the pointer to
		// the allocated memory and not the double pointer
		nestedField.Set(reflect.New(field.FieldType.Elem()))
		fieldInterface = nestedField.Interface()
	}

	fieldSetBytes, ok := fieldInterface.(SetBytesInterface)
	if !ok {
		return fmt.Errorf("field does not satisfy the `SetBytes([]byte)` interface: %T", fieldInterface)
	}

	fieldSetBytes.SetBytes(b)
	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (BytesSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil || (field.FieldType.Kind() == reflect.Pointer && reflect.ValueOf(fieldValue).IsNil()) {
		return nil, nil
	}

	fieldBytes, ok := fieldValue.(BytesInterface)
	if !ok {
		return nil, fmt.Errorf("field does not satisfy the `Bytes() []byte` interface")
	}

	hexStr := hexutil.Encode(fieldBytes.Bytes())
	return hexStr, nil
}
