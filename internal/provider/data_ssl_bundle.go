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
var _ datasource.DataSource = &SSLBundleDataSource{}

type SSLBundleDataSource struct {
	client *porkbun.Client
}

type SSLBundleDataSourceModel struct {
	Domain                  types.String `tfsdk:"domain"`
	IntermediateCertificate types.String `tfsdk:"intermediate_certificate"`
	CertificateChain        types.String `tfsdk:"certificate_chain"`
	PublicKey               types.String `tfsdk:"public_key"`
	PrivateKey              types.String `tfsdk:"private_key"`
}

func NewSSLBundleDataSource() datasource.DataSource {
	return &SSLBundleDataSource{}
}

func (d *SSLBundleDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_ssl_bundle"
}

func (d *SSLBundleDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve the SSL certificate bundle for the domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Your domain.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"intermediate_certificate": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The intermediate certificate.",
			},
			"certificate_chain": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The complete certificate chain.",
			},
			"public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The public key.",
			},
			"private_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The private key.",
			},
		},
	}
}

func (d *SSLBundleDataSource) Configure(
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

func (d *SSLBundleDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var model SSLBundleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := model.Domain.ValueString()

	bundle, err := d.client.SSLBundle(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	model.IntermediateCertificate = types.StringValue(bundle.IntermediateCertificate)
	model.CertificateChain = types.StringValue(bundle.CertificateChain)
	model.PublicKey = types.StringValue(bundle.PublicKey)
	model.PrivateKey = types.StringValue(bundle.PrivateKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
