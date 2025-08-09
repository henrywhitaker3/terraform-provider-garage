// Package provider
package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/henrywhitaker3/terraform-provider-garage/internal/client"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &GarageProvider{}
var _ provider.ProviderWithFunctions = &GarageProvider{}
var _ provider.ProviderWithEphemeralResources = &GarageProvider{}

// GarageProvider defines the provider implementation.
type GarageProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GarageProviderModel describes the provider data model.
type GarageProviderModel struct {
	Host   types.String `tfsdk:"host"`
	Scheme types.String `tfsdk:"scheme"`
	Token  types.String `tfsdk:"token"`
}

func (p *GarageProvider) Metadata(
	ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "garage"
	resp.Version = p.version
}

func (p *GarageProvider) Schema(
	ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Hostname/ip to access the garage api",
				Required:            true,
			},
			"scheme": schema.StringAttribute{
				MarkdownDescription: "The scheme to use, i.e.: http or https",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The token to authenticate with the garage api",
				Required:            true,
			},
		},
	}
}

type setupData struct {
	client *client.Client
}

func (p *GarageProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data GarageProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	scheme := data.Scheme.ValueString()
	if scheme == "" {
		scheme = "https"
	}

	if !slices.Contains([]string{"http", "https"}, scheme) {
		resp.Diagnostics.AddError(
			"invalid scheme value",
			fmt.Sprintf("scheme must be one of: http, https, got %s", scheme),
		)
		return
	}

	setup := setupData{
		client: client.New(
			fmt.Sprintf("%s://%s", scheme, data.Host.ValueString()),
			data.Token.ValueString(),
		),
	}

	resp.DataSourceData = setup
	resp.ResourceData = setup
}

func (p *GarageProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBucketResource,
		NewAccessKeyResource,
		NewPermissionResource,
	}
}

func (p *GarageProvider) EphemeralResources(
	ctx context.Context,
) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *GarageProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *GarageProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GarageProvider{
			version: version,
		}
	}
}
