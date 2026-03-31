package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultBaseURL = "https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app"

var _ provider.Provider = &PulsechecksProvider{}

type PulsechecksProvider struct{}

type PulsechecksProviderModel struct {
	APIURL   types.String `tfsdk:"api_url"`
	APIToken types.String `tfsdk:"api_token"`
}

func New() provider.Provider {
	return &PulsechecksProvider{}
}

func (p *PulsechecksProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pulsechecks"
}

func (p *PulsechecksProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "PulseChecks API base URL. Defaults to the production URL.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Bearer token for API authentication. Can also be set via PULSECHECKS_API_TOKEN env var.",
			},
		},
	}
}

func (p *PulsechecksProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PulsechecksProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := defaultBaseURL
	if !config.APIURL.IsNull() && !config.APIURL.IsUnknown() {
		baseURL = config.APIURL.ValueString()
	}

	token := os.Getenv("PULSECHECKS_API_TOKEN")
	if !config.APIToken.IsNull() && !config.APIToken.IsUnknown() {
		token = config.APIToken.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError("Missing API token", "Set api_token or PULSECHECKS_API_TOKEN env var.")
		return
	}

	client := NewClient(baseURL, token)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PulsechecksProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewCheckResource}
}

func (p *PulsechecksProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewCheckDataSource}
}
