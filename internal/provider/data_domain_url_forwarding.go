// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DomainURLForwardingDataSource{}

type DomainURLForwardingDataSource struct {
	client *porkbun.Client
}

type DomainURLForwardingDataSourceModel struct {
	Domain   types.String   `tfsdk:"domain"`
	Forwards []forwardModel `tfsdk:"forwards"`
}

type forwardModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Subdomain   types.String `tfsdk:"subdomain"`
	Location    types.String `tfsdk:"location"`
	Type        types.String `tfsdk:"type"`
	IncludePath types.Bool   `tfsdk:"include_path"`
	Wildcard    types.Bool   `tfsdk:"wildcard"`
}

func NewDomainURLForwardingDataSource() datasource.DataSource {
	return &DomainURLForwardingDataSource{}
}

func (d *DomainURLForwardingDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_domain_url_forwarding"
}

func (d *DomainURLForwardingDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get URL forwarding for a domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Your domain.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"forwards": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "An array of forwarding records for the domain.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"subdomain": schema.StringAttribute{
							Computed: true,
						},
						"location": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"include_path": schema.BoolAttribute{
							Computed: true,
						},
						"wildcard": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *DomainURLForwardingDataSource) Configure(
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

func (d *DomainURLForwardingDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var model DomainURLForwardingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	forwards, errs := d.client.URLForwards(ctx, domain)
	if errs != nil && len(errs) != 0 {
		for _, err := range errs {
			resp.Diagnostics.AddError("Client Error", err.Error())
		}
		return
	}

	models, diags := forwardsToModels(forwards)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Forwards = models
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func forwardsToModels(
	forwards []*porkbun.URLForward,
) (
	[]forwardModel,
	diag.Diagnostics,
) {
	models := []forwardModel{}
	diags := diag.Diagnostics{}
	for _, forward := range forwards {
		if forward.Type != "temporary" && forward.Type != "permanent" {
			diags.AddWarning("Client Warning", fmt.Sprintf(
				"Expected type of 'temporary' or 'permanent', got: '%s'",
				forward.Type,
			))
		}
		models = append(models, forwardModel{
			ID:          types.Int64Value(*forward.ID),
			Subdomain:   types.StringValue(forward.Subdomain),
			Location:    types.StringValue(forward.Location),
			Type:        types.StringValue(forward.Type),
			IncludePath: types.BoolValue(forward.IncludePath),
			Wildcard:    types.BoolValue(forward.Wildcard),
		})
	}
	return models, diags
}
