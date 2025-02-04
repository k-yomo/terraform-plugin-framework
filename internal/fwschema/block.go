package fwschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Block is the core interface required for implementing Terraform schema
// functionality that structurally holds attributes and blocks. This is
// intended to be the first abstraction of tfsdk.Block functionality into
// data source, provider, and resource specific functionality.
//
// Refer to the internal/fwschema/fwxschema package for optional interfaces
// that define framework-specific functionality, such a plan modification and
// validation.
//
// Note that MaxItems and MinItems support, while defined in the Terraform
// protocol, is intentially not present. Terraform can only perform limited
// static analysis of blocks and errors generated occur before the provider
// is called for configuration validation, which means that practitioners do
// not get all configuration errors at the same time. Provider developers can
// implement validators to achieve the same validation functionality.
type Block interface {
	// Implementations should include the tftypes.AttributePathStepper
	// interface methods for proper path and data handling.
	tftypes.AttributePathStepper

	// Equal should return true if the other block is exactly equivalent.
	Equal(o Block) bool

	// GetDeprecationMessage should return a non-empty string if an attribute
	// is deprecated. This is named differently than DeprecationMessage to
	// prevent a conflict with the tfsdk.Attribute field name.
	GetDeprecationMessage() string

	// GetDescription should return a non-empty string if an attribute
	// has a plaintext description. This is named differently than Description
	// to prevent a conflict with the tfsdk.Attribute field name.
	GetDescription() string

	// GetMarkdownDescription should return a non-empty string if an attribute
	// has a Markdown description. This is named differently than
	// MarkdownDescription to prevent a conflict with the tfsdk.Attribute field
	// name.
	GetMarkdownDescription() string

	// GetNestedObject should return the object underneath the block.
	// For single nesting mode, the NestedBlockObject can be generated from
	// the Block.
	GetNestedObject() NestedBlockObject

	// GetNestingMode should return the nesting mode of a block. This is named
	// differently than NestingMode to prevent a conflict with the tfsdk.Block
	// field name.
	GetNestingMode() BlockNestingMode

	// Type should return the framework type of a block.
	Type() attr.Type
}

// BlocksEqual is a helper function to perform equality testing on two
// Block. Attribute Equal implementations should still compare the concrete
// types in addition to using this helper.
func BlocksEqual(a, b Block) bool {
	if !a.Type().Equal(b.Type()) {
		return false
	}

	if a.GetDeprecationMessage() != b.GetDeprecationMessage() {
		return false
	}

	if a.GetDescription() != b.GetDescription() {
		return false
	}

	if a.GetMarkdownDescription() != b.GetMarkdownDescription() {
		return false
	}

	return true
}
