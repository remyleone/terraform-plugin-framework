package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAttributeModifyPlan(t *testing.T) {
	t.Parallel()

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testCases := map[string]struct {
		attribute    fwschema.Attribute
		req          tfsdk.ModifyAttributePlanRequest
		expectedResp ModifyAttributePlanResponse
	}{
		"no-plan-modifiers": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("testvalue"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"attribute-plan": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					planmodifiers.TestAttrPlanValueModifierTwo{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("MODIFIED_TWO"),
				Private:       testEmptyProviderData,
			},
		},
		"attribute-request-private": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanPrivateModifierGet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
				Private:         testProviderData,
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTATTRONE"),
				Private:       testProviderData,
			},
		},
		"attribute-response-private": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTATTRONE"),
				Private:       testProviderData,
			},
		},
		"attribute-list-nested-private": {
			attribute: tfsdk.Attribute{
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				}),
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"attribute-list-nested-custom": {
			attribute: tfsdk.Attribute{
				Attributes: testtypes.ListNestedAttributesCustomType{
					NestedAttributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
						"nested_attr": {
							Type:     types.StringType,
							Required: true,
							PlanModifiers: tfsdk.AttributePlanModifiers{
								planmodifiers.TestAttrPlanValueModifierOne{},
								planmodifiers.TestAttrPlanValueModifierTwo{},
							},
						},
					}),
				},
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: testtypes.ListNestedAttributesCustomValue{
					List: types.ListValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
				AttributePath: path.Root("test"),
				AttributePlan: testtypes.ListNestedAttributesCustomValue{
					List: types.ListValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
				AttributeState: testtypes.ListNestedAttributesCustomValue{
					List: types.ListValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("MODIFIED_TWO"),
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"attribute-set-nested-private": {
			attribute: tfsdk.Attribute{
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				}),
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"attribute-custom-set-nested": {
			attribute: tfsdk.Attribute{
				Attributes: testtypes.SetNestedAttributesCustomType{
					NestedAttributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
						"nested_attr": {
							Type:     types.StringType,
							Required: true,
							PlanModifiers: tfsdk.AttributePlanModifiers{
								planmodifiers.TestAttrPlanValueModifierOne{},
								planmodifiers.TestAttrPlanValueModifierTwo{},
							},
						},
					}),
				},
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: testtypes.SetNestedAttributesCustomValue{
					Set: types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
				AttributePath: path.Root("test"),
				AttributePlan: testtypes.SetNestedAttributesCustomValue{
					Set: types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
				AttributeState: testtypes.SetNestedAttributesCustomValue{
					Set: types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("MODIFIED_TWO"),
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"attribute-set-nested-usestateforunknown": {
			attribute: tfsdk.Attribute{
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"nested_computed": {
						Type:     types.StringType,
						Computed: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
					"nested_required": {
						Type:     types.StringType,
						Required: true,
					},
				}),
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"attribute-map-nested-private": {
			attribute: tfsdk.Attribute{
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				}),
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"attribute-custom-map-nested": {
			attribute: tfsdk.Attribute{
				Attributes: testtypes.MapNestedAttributesCustomType{
					NestedAttributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
						"nested_attr": {
							Type:     types.StringType,
							Required: true,
							PlanModifiers: tfsdk.AttributePlanModifiers{
								planmodifiers.TestAttrPlanValueModifierOne{},
								planmodifiers.TestAttrPlanValueModifierTwo{},
							},
						},
					}),
				},
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: testtypes.MapNestedAttributesCustomValue{
					Map: types.MapValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						map[string]attr.Value{
							"testkey": types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
				AttributePath: path.Root("test"),
				AttributePlan: testtypes.MapNestedAttributesCustomValue{
					Map: types.MapValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						map[string]attr.Value{
							"testkey": types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
				AttributeState: testtypes.MapNestedAttributesCustomValue{
					Map: types.MapValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
						map[string]attr.Value{
							"testkey": types.ObjectValueMust(
								map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								map[string]attr.Value{
									"nested_attr": types.StringValue("TESTATTRONE"),
								},
							),
						},
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("MODIFIED_TWO"),
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"attribute-single-nested-private": {
			attribute: tfsdk.Attribute{
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"testing": {
						Type:     types.StringType,
						Optional: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				}),
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				Private: testProviderData,
			},
		},
		"attribute-custom-single-nested": {
			attribute: tfsdk.Attribute{
				Attributes: testtypes.SingleNestedAttributesCustomType{
					NestedAttributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"testing": {
							Type:     types.StringType,
							Optional: true,
							PlanModifiers: tfsdk.AttributePlanModifiers{
								planmodifiers.TestAttrPlanValueModifierOne{},
								planmodifiers.TestAttrPlanValueModifierTwo{},
							},
						},
					}),
				},
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: testtypes.SingleNestedAttributesCustomValue{
					Object: types.ObjectValueMust(
						map[string]attr.Type{
							"testing": types.StringType,
						},
						map[string]attr.Value{
							"testing": types.StringValue("TESTATTRONE"),
						},
					),
				},
				AttributePath: path.Root("test"),
				AttributePlan: testtypes.SingleNestedAttributesCustomValue{
					Object: types.ObjectValueMust(
						map[string]attr.Type{
							"testing": types.StringType,
						},
						map[string]attr.Value{
							"testing": types.StringValue("TESTATTRONE"),
						},
					),
				},
				AttributeState: testtypes.SingleNestedAttributesCustomValue{
					Object: types.ObjectValueMust(
						map[string]attr.Type{
							"testing": types.StringType,
						},
						map[string]attr.Value{
							"testing": types.StringValue("TESTATTRONE"),
						},
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("MODIFIED_TWO"),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"requires-replacement": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("newtestvalue"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("newtestvalue"),
				AttributeState:  types.StringValue("testvalue"),
				// resource.RequiresReplace() requires non-null plan
				// and state.
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("newtestvalue"),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"requires-replacement-passthrough": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					resource.RequiresReplace(),
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
				// resource.RequiresReplace() requires non-null plan
				// and state.
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestAttrPlanValueModifierOne{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestAttrPlanValueModifierOne{},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTATTRTWO"),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"requires-replacement-unset": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestRequiresReplaceFalseModifier{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("testvalue"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				Private:       testEmptyProviderData,
			},
		},
		"warnings": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTDIAG"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTDIAG"),
				AttributeState:  types.StringValue("TESTDIAG"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTDIAG"),
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				Private: testEmptyProviderData,
			},
		},
		"error": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTDIAG"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTDIAG"),
				AttributeState:  types.StringValue("TESTDIAG"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTDIAG"),
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Private: testEmptyProviderData,
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got := ModifyAttributePlanResponse{
				AttributePlan: tc.req.AttributePlan,
				Private:       tc.req.Private,
			}

			AttributeModifyPlan(ctx, tc.attribute, tc.req, &got)

			if diff := cmp.Diff(tc.expectedResp, got, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("Unexpected response (-wanted, +got): %s", diff)
			}
		})
	}
}
