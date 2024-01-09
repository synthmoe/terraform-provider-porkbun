// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &DomainNameServersResource{}

type DomainNameServersResource struct {
	deleteNameServers bool
	client            *porkbun.Client
}

type DomainNameServersResourceModel struct {
	Domain types.String `tfsdk:"domain"`
	NS     types.List   `tfsdk:"ns"`
}

func NewDomainNameServersResource() resource.Resource {
	return &DomainNameServersResource{}
}

func (r *DomainNameServersResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_domain_name_servers"
}

func (r *DomainNameServersResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Update the name servers for your domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Your domain.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ns": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "An array of name servers that you would like to update your domain with.",
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 253),
					),
				},
			},
		},
	}
}

func (r *DomainNameServersResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*PorkbunProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *PorkbunProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.deleteNameServers = data.DeleteNameServers
	r.client = data.Client
}

func (r *DomainNameServersResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var model DomainNameServersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	var ns []string
	resp.Diagnostics.Append(model.NS.ElementsAs(ctx, &ns, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNameServers(ctx, domain, ns)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DomainNameServersResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var model DomainNameServersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	servers, err := r.client.NameServers(ctx, domain)
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

func (r *DomainNameServersResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var model DomainNameServersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	var ns []string
	resp.Diagnostics.Append(model.NS.ElementsAs(ctx, &ns, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNameServers(ctx, domain, ns)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DomainNameServersResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	if !r.deleteNameServers {
		return
	}

	var model DomainNameServersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	ns := []string{}

	err := r.client.UpdateNameServers(ctx, domain, ns)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
}
