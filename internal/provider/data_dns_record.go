// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DNSRecrodDataSource{}

type DNSRecrodDataSource struct {
	client *porkbun.Client
}

type DNSRecrodDataSourceModel struct {
	Domain    types.String  `tfsdk:"domain"`
	Subdomain types.String  `tfsdk:"subdomain"`
	Type      types.String  `tfsdk:"type"`
	ID        types.Int64   `tfsdk:"id"`
	Records   []recordModel `tfsdk:"records"`
}

type recordModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Subdomain types.String `tfsdk:"subdomain"`
	Type      types.String `tfsdk:"type"`
	Content   types.String `tfsdk:"content"`
	TTL       types.Int64  `tfsdk:"ttl"`
	Priority  types.Int64  `tfsdk:"priority"`
	Notes     types.String `tfsdk:"notes"`
}

func NewDNSRecrodDataSource() datasource.DataSource {
	return &DNSRecrodDataSource{}
}

func (d *DNSRecrodDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (d *DNSRecrodDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve all editable DNS records associated with a domain, subdomain and type " +
			"or a single record for a particular record ID.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Your domain.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"subdomain": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Your subdomain. Requires type to be set.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("type"),
					}...),
				},
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Record type. Mutually exclusive with id.",
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
			"id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Record ID. Mutually exclusive with type.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("type"),
					}...),
				},
			},
			"records": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "An array of DNS records.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"subdomain": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"content": schema.StringAttribute{
							Computed: true,
						},
						"ttl": schema.Int64Attribute{
							Computed: true,
						},
						"priority": schema.Int64Attribute{
							Computed: true,
						},
						"notes": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *DNSRecrodDataSource) Configure(
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

func (d *DNSRecrodDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var model DNSRecrodDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()
	subdomain := model.Subdomain.ValueString()
	type_ := model.Type.ValueString()
	id := model.ID.ValueInt64()

	var records []*porkbun.DNSRecord
	var errs []error
	if type_ != "" {
		records, errs = d.client.DNSRecordsByTypeName(ctx, domain, type_, subdomain)
	} else {
		records, errs = d.client.DNSRecords(ctx, domain, &id)
	}
	if errs != nil && len(errs) != 0 {
		for _, err := range errs {
			resp.Diagnostics.AddError("Client Error", err.Error())
		}
		return
	}

	models := []recordModel{}
	for _, record := range records {
		id := types.Int64Null()
		if record.ID != nil {
			id = types.Int64Value(*record.ID)
		}
		priority := types.Int64Null()
		if record.Priority != nil {
			priority = types.Int64Value(*record.Priority)
		}
		models = append(models, recordModel{
			ID:        id,
			Subdomain: types.StringValue(record.Subdomain),
			Type:      types.StringValue(record.Type),
			Content:   types.StringValue(record.Content),
			TTL:       types.Int64Value(record.TTL),
			Priority:  priority,
			Notes:     types.StringValue(record.Notes),
		})
	}
	model.Records = models
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
