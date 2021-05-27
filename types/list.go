package types

import (
	"context"
	"fmt"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ListType is an AttributeType representing a list of values. All values must
// be of the same type, which the provider must specify as the ElemType
// property.
type ListType struct {
	ElemType tfsdk.AttributeType
}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (l ListType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.List{
		ElementType: l.ElemType.TerraformType(ctx),
	}
}

// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (l ListType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	if !in.IsKnown() {
		return List{
			Unknown: true,
		}, nil
	}
	if in.IsNull() {
		return List{
			Null: true,
		}, nil
	}
	val := []tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	elems := make([]tfsdk.AttributeValue, 0, len(val))
	for _, elem := range val {
		av, err := l.ElemType.ValueFromTerraform(ctx, elem)
		if err != nil {
			return nil, err
		}
		elems = append(elems, av)
	}
	return List{
		Elems:    elems,
		ElemType: l.TerraformType(ctx),
	}, nil
}

// List represents a list of AttributeValues, all of the same type, indicated
// by ElemType.
type List struct {
	// Unknown will be set to true if the entire list is an unknown value.
	// If only some of the elements in the list are unknown, their known or
	// unknown status will be represented however that AttributeValue
	// surfaces that information. The List's Unknown property only tracks
	// if the number of elements in a List is known, not whether the
	// elements that are in the list are known.
	Unknown bool

	// Null will be set to true if the list is null, either because it was
	// omitted from the configuration, state, or plan, or because it was
	// explicitly set to null.
	Null bool

	// Elems are the elements in the list.
	Elems []tfsdk.AttributeValue

	// ElemType is the tftypes.Type of the elements in the list. All
	// elements in the list must be of this type.
	ElemType tftypes.Type
}

// ElementsAs populates `target` with the elements of the List, throwing an
// error if the elements cannot be stored in `target`.
func (l List) ElementsAs(ctx context.Context, target interface{}, allowUnhandled bool) error {
	// we need a tftypes.Value for this List to be able to use it with our
	// reflection code
	values := make([]tftypes.Value, 0, len(l.Elems))
	for pos, elem := range l.Elems {
		val, err := elem.ToTerraformValue(ctx)
		if err != nil {
			return fmt.Errorf("error getting Terraform value for element %d: %w", pos, err)
		}
		err = tftypes.ValidateValue(l.ElemType, val)
		if err != nil {
			return fmt.Errorf("error using created Terraform value for element %d: %w", pos, err)
		}
		values = append(values, tftypes.NewValue(l.ElemType, val))
	}
	return reflect.Into(ctx, tftypes.NewValue(tftypes.List{
		ElementType: l.ElemType,
	}, values), target, reflect.Options{
		UnhandledNullAsEmpty:    allowUnhandled,
		UnhandledUnknownAsEmpty: allowUnhandled,
	}, tftypes.NewAttributePath())
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (l List) ToTerraformValue(ctx context.Context) (interface{}, error) {
	if l.Unknown {
		return tftypes.UnknownValue, nil
	}
	if l.Null {
		return nil, nil
	}
	vals := make([]tftypes.Value, 0, len(l.Elems))
	for _, elem := range l.Elems {
		val, err := elem.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		err = tftypes.ValidateValue(l.ElemType, val)
		if err != nil {
			return nil, err
		}
		vals = append(vals, tftypes.NewValue(l.ElemType, val))
	}
	return vals, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (l List) Equal(o tfsdk.AttributeValue) bool {
	other, ok := o.(List)
	if !ok {
		return false
	}
	if l.Unknown != other.Unknown {
		return false
	}
	if l.Null != other.Null {
		return false
	}
	if !l.ElemType.Is(other.ElemType) {
		return false
	}
	if len(l.Elems) != len(other.Elems) {
		return false
	}
	for pos, lElem := range l.Elems {
		oElem := other.Elems[pos]
		if !lElem.Equal(oElem) {
			return false
		}
	}
	return true
}