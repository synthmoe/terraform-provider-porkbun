// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"

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
var _ resource.Resource = &DomainURLForwardResource{}

type DomainURLForwardResource struct {
	client *porkbun.Client
}

type DomainURLForwardResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Domain      types.String `tfsdk:"domain"`
	Subdomain   types.String `tfsdk:"subdomain"`
	Location    types.String `tfsdk:"location"`
	Type        types.String `tfsdk:"type"`
	IncludePath types.Bool   `tfsdk:"include_path"`
	Wildcard    types.Bool   `tfsdk:"wildcard"`
}

func NewDomainURLForwardResource() resource.Resource {
	return &DomainURLForwardResource{}
}

func (r *DomainURLForwardResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_domain_url_forward"
}

func (r *DomainURLForwardResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "URL forward for a domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
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
			"subdomain": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "A subdomain that you would like to add email forwarding for. " +
					"Leave this blank to forward the root domain.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"location": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Where you'd like to forward the domain to.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2000),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of forward. Valid types are: temporary or permanent.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"temporary",
						"permanent",
					),
				},
			},
			"include_path": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether or not to include the URI path in the redirection.",
			},
			"wildcard": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether or not to forward all subdomains of the domain.",
			},
		},
	}
}

func (r *DomainURLForwardResource) Configure(
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

	r.client = data.Client
}

func (r *DomainURLForwardResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var model DomainURLForwardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	subdomain := model.Subdomain.ValueString()
	location := model.Location.ValueString()
	type_ := model.Type.ValueString()
	includePath := model.IncludePath.ValueBool()
	wildcard := model.Wildcard.ValueBool()

	err := r.client.AddURLForward(ctx, domain, &porkbun.URLForward{
		Subdomain:   subdomain,
		Location:    location,
		Type:        type_,
		IncludePath: includePath,
		Wildcard:    wildcard,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	forwards, errs := r.client.URLForwards(ctx, domain)
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

	id := int64(0)
	for _, forward := range models {
		if model.Subdomain.Equal(forward.Subdomain) &&
			model.Location.Equal(forward.Location) &&
			model.Type.Equal(forward.Type) &&
			model.IncludePath.Equal(forward.IncludePath) &&
			model.Wildcard.Equal(forward.Wildcard) {
			id = forward.ID.ValueInt64()
		}
	}
	if id == 0 {
		resp.Diagnostics.AddError("Client Error", "Failed to obtain ID of newly created forward.")
		return
	}

	model.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DomainURLForwardResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var model DomainURLForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	forwards, errs := r.client.URLForwards(ctx, domain)
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

	found := false
	for _, forward := range models {
		if model.ID.Equal(forward.ID) {
			found = true
			model.Subdomain = forward.Subdomain
			model.Location = forward.Location
			model.Type = forward.Type
			model.IncludePath = forward.IncludePath
			model.Wildcard = forward.Wildcard
		}
	}
	if !found {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf(
			"Failed to find url forward with the id '%d'.",
			model.ID,
		))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DomainURLForwardResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var model DomainURLForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	id := model.ID.ValueInt64()

	err := r.client.DeleteURLForward(ctx, domain, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain = model.Domain.ValueString()
	subdomain := model.Subdomain.ValueString()
	location := model.Location.ValueString()
	type_ := model.Type.ValueString()
	includePath := model.IncludePath.ValueBool()
	wildcard := model.Wildcard.ValueBool()

	err = r.client.AddURLForward(ctx, domain, &porkbun.URLForward{
		Subdomain:   subdomain,
		Location:    location,
		Type:        type_,
		IncludePath: includePath,
		Wildcard:    wildcard,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	forwards, errs := r.client.URLForwards(ctx, domain)
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

	id = int64(0)
	for _, forward := range models {
		if model.Subdomain.Equal(forward.Subdomain) &&
			model.Location.Equal(forward.Location) &&
			model.Type.Equal(forward.Type) &&
			model.IncludePath.Equal(forward.IncludePath) &&
			model.Wildcard.Equal(forward.Wildcard) {
			id = forward.ID.ValueInt64()
		}
	}
	if id == 0 {
		resp.Diagnostics.AddError("Client Error", "Failed to obtain ID of newly created forward.")
		return
	}

	model.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DomainURLForwardResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var model DomainURLForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	id := model.ID.ValueInt64()

	err := r.client.DeleteURLForward(ctx, domain, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
}
