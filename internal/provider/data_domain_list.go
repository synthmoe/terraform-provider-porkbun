// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DomainListDataSource{}

type DomainListDataSource struct {
	client *porkbun.Client
}

type DomainListDataSourceModel struct {
	Domains []domainModel `tfsdk:"domains"`
}

type domainModel struct {
	Domain       types.String `tfsdk:"domain"`
	Status       types.String `tfsdk:"status"`
	TLD          types.String `tfsdk:"tld"`
	CreateDate   types.String `tfsdk:"create_date"`
	ExpireDate   types.String `tfsdk:"expire_date"`
	SecurityLock types.Bool   `tfsdk:"security_lock"`
	WhoisPrivacy types.Bool   `tfsdk:"whois_privacy"`
	AutoRenew    types.Bool   `tfsdk:"auto_renew"`
	NotLocal     types.Bool   `tfsdk:"not_local"`
}

func NewDomainListDataSource() datasource.DataSource {
	return &DomainListDataSource{}
}

func (d *DomainListDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_domain_list"
}

func (d *DomainListDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get all domain names in account. Domains are returned in chunks of 1000.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "An array of domains in the account.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"tld": schema.StringAttribute{
							Computed: true,
						},
						"create_date": schema.StringAttribute{
							Computed: true,
						},
						"expire_date": schema.StringAttribute{
							Computed: true,
						},
						"security_lock": schema.BoolAttribute{
							Computed: true,
						},
						"whois_privacy": schema.BoolAttribute{
							Computed: true,
						},
						"auto_renew": schema.BoolAttribute{
							Computed: true,
						},
						"not_local": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *DomainListDataSource) Configure(
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

func (d *DomainListDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var model DomainListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	models := []domainModel{}
	for start := int64(0); ; start += 1000 {
		domains, errs := d.client.DomainList(ctx, start)
		if errs != nil && len(errs) != 0 {
			for _, err := range errs {
				resp.Diagnostics.AddError("Client Error", err.Error())
			}
			break
		}
		for _, domain := range domains {
			models = append(models, domainModel{
				Domain:       types.StringValue(domain.Domain),
				Status:       types.StringValue(domain.Status),
				TLD:          types.StringValue(domain.TLD),
				CreateDate:   types.StringValue(domain.CreateDate),
				ExpireDate:   types.StringValue(domain.ExpireDate),
				SecurityLock: types.BoolValue(domain.SecurityLock),
				WhoisPrivacy: types.BoolValue(domain.WhoisPrivacy),
				AutoRenew:    types.BoolValue(domain.AutoRenew),
				NotLocal:     types.BoolValue(domain.NotLocal),
			})
		}
		if len(domains) < 1000 {
			break
		}
	}

	model.Domains = models
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
