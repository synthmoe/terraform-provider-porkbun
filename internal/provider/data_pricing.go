// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PricingDataSource{}

type PricingDataSource struct {
	client *porkbun.Client
}

type PricingDataSourceModel struct {
	Pricing map[string]pricingModel `tfsdk:"pricing"`
}

type pricingModel struct {
	Registration types.Float64 `tfsdk:"registration"`
	Renewal      types.Float64 `tfsdk:"renewal"`
	Transfer     types.Float64 `tfsdk:"transfer"`
	SpecialType  types.String  `tfsdk:"special_type"`
}

func NewPricingDataSource() datasource.DataSource {
	return &PricingDataSource{}
}

func (d *PricingDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_pricing"
}

func (d *PricingDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Check default domain pricing information for all supported TLDs.",
		Attributes: map[string]schema.Attribute{
			"pricing": schema.MapNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Objects with default pricing for the registration, renewal and transfer of each supported TLD.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"registration": schema.Float64Attribute{
							Computed: true,
						},
						"renewal": schema.Float64Attribute{
							Computed: true,
						},
						"transfer": schema.Float64Attribute{
							Computed: true,
						},
						"special_type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *PricingDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*PorkbunProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *PorkbunProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = data.Client
}

func (d *PricingDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var model PricingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pricing, err := d.client.Pricing(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	models := map[string]pricingModel{}
	for k, v := range pricing {
		registration, err := strconv.ParseFloat(strings.ReplaceAll(v.Registration, ",", ""), 64)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf(
				"Failed to parse registration as an integer with the following error: '%s'.",
				err.Error(),
			))
			continue
		}
		renewal, err := strconv.ParseFloat(strings.ReplaceAll(v.Renewal, ",", ""), 64)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf(
				"Failed to parse renewal as an integer with the following error: '%s'.",
				err.Error(),
			))
			continue
		}
		transfer, err := strconv.ParseFloat(strings.ReplaceAll(v.Transfer, ",", ""), 64)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf(
				"Failed to parse transfer as an integer with the following error: '%s'.",
				err.Error(),
			))
			continue
		}
		var specialType types.String
		if v.SpecialType == "" {
			specialType = types.StringNull()
		} else {
			specialType = types.StringValue(v.SpecialType)
		}
		models[k] = pricingModel{
			Registration: types.Float64Value(registration),
			Renewal:      types.Float64Value(renewal),
			Transfer:     types.Float64Value(transfer),
			SpecialType:  specialType,
		}
	}

	model.Pricing = models
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
