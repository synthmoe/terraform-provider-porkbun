// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSRecordResource{}

type DNSRecordResource struct {
	client *porkbun.Client
}

type DNSRecordResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Domain    types.String `tfsdk:"domain"`
	Subdomain types.String `tfsdk:"subdomain"`
	Type      types.String `tfsdk:"type"`
	Content   types.String `tfsdk:"content"`
	TTL       types.Int64  `tfsdk:"ttl"`
	Priority  types.Int64  `tfsdk:"priority"`
	Notes     types.String `tfsdk:"notes"`
}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

func (r *DNSRecordResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create a DNS record.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the record created.",
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
				MarkdownDescription: "The subdomain for the record being created, not including the domain itself. " +
					"Leave blank to create a record on the root domain. Use * to create a wildcard record.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of record being created. Valid types are: A, MX, CNAME, ALIAS, TXT, NS, AAAA, SRV, TLSA, CAA.",
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(
						"A",
						"MX",
						"CNAME",
						"ALIAS",
						"TXT",
						"NS",
						"AAAA",
						"SRV",
						"TLSA",
						"CAA",
					),
				},
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The answer content for the record.",
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(600),
				MarkdownDescription: "The time to live in seconds for the record. " +
					"The minimum and the default is 600 seconds.",
				Validators: []validator.Int64{
					int64validator.Between(600, int64(math.Pow(2, 31)-1)),
				},
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The priority of the record for those that support it.",
				Validators: []validator.Int64{
					int64validator.Between(0, int64(math.Pow(2, 16)-1)),
				},
			},
			"notes": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Currently doesn't do anything. :(",
			},
		},
	}
}

func (r *DNSRecordResource) Configure(
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

func (r *DNSRecordResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var model DNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	subdomain := model.Subdomain.ValueString()
	type_ := model.Type.ValueString()
	content := model.Content.ValueString()
	ttl := model.TTL.ValueInt64()
	var priority *int64
	if !model.Priority.IsNull() {
		priority = &[]int64{model.Priority.ValueInt64()}[0]
	}
	notes := model.Notes.ValueString()

	id, err := r.client.CreateDNSRecord(ctx, domain, &porkbun.DNSRecord{
		Subdomain: subdomain,
		Type:      type_,
		Content:   content,
		TTL:       ttl,
		Priority:  priority,
		Notes:     notes,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	model.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DNSRecordResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var model DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	id := model.ID.ValueInt64()

	records, errs := r.client.DNSRecords(ctx, domain, &id)
	if errs != nil && len(errs) != 0 {
		for _, err := range errs {
			resp.Diagnostics.AddError("Client Error", err.Error())
		}
		return
	}
	if len(records) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf(
			"Expected to receive 1 record, got: %d.",
			len(records),
		))
	}

	record := records[0]

	priority := types.Int64Null()
	if record.Priority != nil {
		priority = types.Int64Value(*record.Priority)
	}

	model.Subdomain = types.StringValue(record.Subdomain)
	model.Type = types.StringValue(record.Type)
	model.Content = types.StringValue(record.Content)
	model.TTL = types.Int64Value(record.TTL)
	model.Priority = priority
	model.Notes = types.StringValue(record.Notes)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DNSRecordResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var model DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	id := model.ID.ValueInt64()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subdomain := model.Subdomain.ValueString()
	type_ := model.Type.ValueString()
	content := model.Content.ValueString()
	ttl := model.TTL.ValueInt64()
	var priority *int64
	if !model.Priority.IsNull() {
		priority = &[]int64{model.Priority.ValueInt64()}[0]
	}
	notes := model.Notes.ValueString()

	err := r.client.EditDNSRecord(ctx, domain, &porkbun.DNSRecord{
		ID:        &id,
		Subdomain: subdomain,
		Type:      type_,
		Content:   content,
		TTL:       ttl,
		Priority:  priority,
		Notes:     notes,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	model.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DNSRecordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var model DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	id := model.ID.ValueInt64()

	err := r.client.DeleteDNSRecord(ctx, domain, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
}
