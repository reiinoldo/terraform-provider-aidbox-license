// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/url"
	"os"

	"github.com/reiinoldo/terraform-provider-aidbox-license/internal/aidbox"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AidboxProvider satisfies various provider interfaces.
var _ provider.Provider = &AidboxProvider{}
var _ provider.ProviderWithFunctions = &AidboxProvider{}

type AidboxProvider struct {
	version string
}

type AidboxProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

type Client interface {
	CreateLicense(ctx context.Context, name, product, licenseType string) (aidbox.LicenseResponse, error)
	GetLicense(ctx context.Context, licenseID string) (aidbox.LicenseResponse, error)
	DeleteLicense(ctx context.Context, licenseID string) error
}

type ProviderData struct {
	Client Client
}

func (p *AidboxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aidbox"
	resp.Version = p.version
}

func (p *AidboxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Aidbox licenses through the Aidbox RPC API. This provider supports Terraform and OpenTofu.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Aidbox RPC API endpoint. Defaults to `https://aidbox.app/rpc`; can also be set with `AIDBOX_ENDPOINT`.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Aidbox API token. Can also be set with the `AIDBOX_API_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AidboxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AidboxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsUnknown() || data.Token.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Provider Configuration",
			"The Aidbox endpoint and token must be known before the provider can be configured.",
		)
		return
	}

	endpoint := data.Endpoint.ValueString()
	if data.Endpoint.IsNull() || endpoint == "" {
		endpoint = os.Getenv("AIDBOX_ENDPOINT")
		if endpoint == "" {
			endpoint = "https://aidbox.app/rpc"
		}
	}
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil || (parsedEndpoint.Scheme != "http" && parsedEndpoint.Scheme != "https") || parsedEndpoint.Host == "" {
		resp.Diagnostics.AddError(
			"Invalid Aidbox Endpoint",
			"The endpoint must be an absolute HTTP or HTTPS URL.",
		)
		return
	}

	if data.Token.IsNull() || data.Token.IsUnknown() || data.Token.ValueString() == "" {
		tokenEnv := os.Getenv("AIDBOX_API_TOKEN")
		if tokenEnv != "" {
			data.Token = types.StringValue(tokenEnv)
		} else {
			resp.Diagnostics.AddError(
				"No API Token Provided",
				"Please provide a 'token' in the provider configuration or through the 'AIDBOX_API_TOKEN' environment variable.",
			)
			return
		}
	}

	resp.ResourceData = &ProviderData{
		Client: aidbox.NewClient(endpoint, data.Token.ValueString()),
	}
}

func (p *AidboxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLicenseResource,
	}
}

func (p *AidboxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *AidboxProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AidboxProvider{
			version: version,
		}
	}
}
