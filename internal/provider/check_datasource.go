package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CheckDataSource{}

type CheckDataSource struct {
	client *ApiClient
}

type CheckDataSourceModel struct {
	TeamID        types.String `tfsdk:"team_id"`
	CheckID       types.String `tfsdk:"check_id"`
	Name          types.String `tfsdk:"name"`
	CheckType     types.String `tfsdk:"check_type"`
	PeriodSeconds types.Int64  `tfsdk:"period_seconds"`
	Schedule      types.String `tfsdk:"schedule"`
	GraceSeconds  types.Int64  `tfsdk:"grace_seconds"`
	Token         types.String `tfsdk:"token"`
	Status        types.String `tfsdk:"status"`
}

func NewCheckDataSource() datasource.DataSource {
	return &CheckDataSource{}
}

func (d *CheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

func (d *CheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a PulseChecks check by ID.",
		Attributes: map[string]schema.Attribute{
			"team_id":        schema.StringAttribute{Required: true},
			"check_id":       schema.StringAttribute{Required: true},
			"name":           schema.StringAttribute{Computed: true},
			"check_type":     schema.StringAttribute{Computed: true},
			"period_seconds": schema.Int64Attribute{Computed: true},
			"schedule":       schema.StringAttribute{Computed: true},
			"grace_seconds":  schema.Int64Attribute{Computed: true},
			"token":          schema.StringAttribute{Computed: true, Sensitive: true},
			"status":         schema.StringAttribute{Computed: true},
		},
	}
}

func (d *CheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ApiClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *ApiClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *CheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CheckDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := d.client.GetCheck(state.TeamID.ValueString(), state.CheckID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read failed", err.Error())
		return
	}
	if check == nil {
		resp.Diagnostics.AddError("Not found", "Check not found")
		return
	}

	state.Name = types.StringValue(check.Name)
	state.CheckType = types.StringValue(check.CheckType)
	state.PeriodSeconds = types.Int64Value(int64(check.PeriodSeconds))
	state.Schedule = types.StringValue(check.Schedule)
	state.GraceSeconds = types.Int64Value(int64(check.GraceSeconds))
	state.Token = types.StringValue(check.Token)
	state.Status = types.StringValue(check.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
