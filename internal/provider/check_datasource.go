package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CheckDataSource{}

type CheckDataSource struct {
	client *Client
}

func NewCheckDataSource() datasource.DataSource {
	return &CheckDataSource{}
}

type CheckDataSourceModel struct {
	TeamID             types.String `tfsdk:"team_id"`
	CheckID            types.String `tfsdk:"check_id"`
	Name               types.String `tfsdk:"name"`
	CheckType          types.String `tfsdk:"check_type"`
	PeriodSeconds      types.Int64  `tfsdk:"period_seconds"`
	GraceSeconds       types.Int64  `tfsdk:"grace_seconds"`
	URL                types.String `tfsdk:"url"`
	ExpectedStatusCode types.Int64  `tfsdk:"expected_status_code"`
	ExpectedString     types.String `tfsdk:"expected_string"`
	FailureThreshold   types.Int64  `tfsdk:"failure_threshold"`
	Token              types.String `tfsdk:"token"`
}

func (d *CheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

func (d *CheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"team_id":              schema.StringAttribute{Required: true},
			"check_id":             schema.StringAttribute{Required: true},
			"name":                 schema.StringAttribute{Computed: true},
			"check_type":           schema.StringAttribute{Computed: true},
			"period_seconds":       schema.Int64Attribute{Computed: true},
			"grace_seconds":        schema.Int64Attribute{Computed: true},
			"url":                  schema.StringAttribute{Computed: true},
			"expected_status_code": schema.Int64Attribute{Computed: true},
			"expected_string":      schema.StringAttribute{Computed: true},
			"failure_threshold":    schema.Int64Attribute{Computed: true},
			"token":                schema.StringAttribute{Computed: true, Sensitive: true},
		},
	}
}

func (d *CheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *CheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CheckDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := d.client.GetCheck(ctx, state.TeamID.ValueString(), state.CheckID.ValueString())
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
	state.PeriodSeconds = types.Int64Value(check.PeriodSeconds)
	state.GraceSeconds = types.Int64Value(check.GraceSeconds)
	state.Token = types.StringValue(check.Token)

	if check.URL != nil {
		state.URL = types.StringValue(*check.URL)
	} else {
		state.URL = types.StringNull()
	}
	if check.ExpectedStatusCode != nil {
		state.ExpectedStatusCode = types.Int64Value(*check.ExpectedStatusCode)
	} else {
		state.ExpectedStatusCode = types.Int64Null()
	}
	if check.ExpectedString != nil {
		state.ExpectedString = types.StringValue(*check.ExpectedString)
	} else {
		state.ExpectedString = types.StringNull()
	}
	if check.FailureThreshold != nil {
		state.FailureThreshold = types.Int64Value(*check.FailureThreshold)
	} else {
		state.FailureThreshold = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
