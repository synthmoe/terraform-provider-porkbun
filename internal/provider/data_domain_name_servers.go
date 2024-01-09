// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DomainNameServersDataSource{}

type DomainNameServersDataSource struct {
	client *porkbun.Client
}

type DomainNameServersDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`
	NS     types.List   `tfsdk:"ns"`
}

func NewDomainNameServersDataSource() datasource.DataSource {
	return &DomainNameServersDataSource{}
}

func (d *DomainNameServersDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_domain_name_servers"
}

func (d *DomainNameServersDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the authoritative name servers listed at the registry for your domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Your domain.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"ns": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "An array of name server host names.",
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *DomainNameServersDataSource) Configure(
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

func (d *DomainNameServersDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var model DomainNameServersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	servers, err := d.client.NameServers(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	ns, diags := types.ListValueFrom(ctx, types.StringType, servers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.NS = ns
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
