// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package provider

import (
	"context"
	"net/http"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/synthmoe/terraform-provider-porkbun/internal/porkbun"
)

// Ensure PorkbunProvider satisfies various provider interfaces.
var _ provider.Provider = &PorkbunProvider{}

var apiKeyRE = regexp.MustCompile(`^pk1_[0-9a-f]{64}$`)
var secretAPIKeyRE = regexp.MustCompile(`^sk1_[0-9a-f]{64}$`)

type PorkbunProvider struct {
	version string
}

type PorkbunProviderModel struct {
	APIKey            types.String `tfsdk:"api_key"`
	SecretAPIKey      types.String `tfsdk:"secret_api_key"`
	ForceIPv4         types.Bool   `tfsdk:"force_ipv4"`
	DeleteNameServers types.Bool   `tfsdk:"delete_name_servers"`
}

type PorkbunProviderData struct {
	DeleteNameServers bool
	Client            *porkbun.Client
}

func (p *PorkbunProvider) Metadata(
	_ context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "porkbun"
	resp.Version = p.version
}

func (p *PorkbunProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Your API key. Overrides PORKBUN_API_KEY environment variable.",
				Optional:            true,
				Validators: []validator.String{stringvalidator.RegexMatches(
					apiKeyRE,
					"must consist of 'pk1_' prefix followed by 64 digit lowercase hexadecimal number",
				)},
			},
			"secret_api_key": schema.StringAttribute{
				MarkdownDescription: "Your secret API key. Overrides PORKBUN_SECRET_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{stringvalidator.RegexMatches(
					secretAPIKeyRE,
					"must consist of 'sk1_' prefix followed by 64 digit lowercase hexadecimal number",
				)},
			},
			"force_ipv4": schema.BoolAttribute{
				MarkdownDescription: "Force the use of IPv4 via the dedicated IPv4 hostname *api-ipv4.porkbun.com* instead of the default *porkbun.com*.",
				Optional:            true,
			},
			"delete_name_servers": schema.BoolAttribute{
				MarkdownDescription: "Delete name servers on terraform destroy by updating them to an empty list. Disabled by default.",
				Optional:            true,
			},
		},
	}
}

func (p *PorkbunProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var model PorkbunProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("PORKBUN_API_KEY")
	if !model.APIKey.IsNull() {
		apiKey = model.APIKey.ValueString()
	}
	if !apiKeyRE.Match([]byte(apiKey)) {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Invalid Porkbun API Key",
			"API key must consist of 'pk1_' prefix followed by a 64 digit lowercase hexadecimal number, got: "+apiKey,
		)
	}

	secretAPIKey := os.Getenv("PORKBUN_SECRET_API_KEY")
	if !model.SecretAPIKey.IsNull() {
		secretAPIKey = model.SecretAPIKey.ValueString()
	}
	if !secretAPIKeyRE.Match([]byte(secretAPIKey)) {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret_api_key"),
			"Invalid Porkbun Secret API Key",
			"Secret API key must consist of 'sk1_' prefix followed by a 64 digit lowercase hexadecimal number, got: "+secretAPIKey,
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	forceIPv4 := model.ForceIPv4.ValueBool()
	deleteNameServers := model.DeleteNameServers.ValueBool()

	client := porkbun.NewClient(http.DefaultClient, apiKey, secretAPIKey, forceIPv4)

	// Try authenticating with supplied keys.
	_, err := client.Ping(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	data := PorkbunProviderData{
		DeleteNameServers: deleteNameServers,
		Client:            client,
	}
	resp.DataSourceData = &data
	resp.ResourceData = &data
}

func (p *PorkbunProvider) DataSources(
	_ context.Context,
) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPricingDataSource,
		NewDomainListDataSource,
		NewDomainNameServersDataSource,
		NewDomainURLForwardingDataSource,
		NewDNSRecrodDataSource,
		NewSSLBundleDataSource,
	}
}

func (p *PorkbunProvider) Resources(
	_ context.Context,
) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainNameServersResource,
		NewDomainURLForwardResource,
		NewDNSRecordResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PorkbunProvider{
			version: version,
		}
	}
}
