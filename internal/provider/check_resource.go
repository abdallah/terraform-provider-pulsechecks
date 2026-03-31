package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CheckResource{}

type CheckResource struct {
	client *Client
}

func NewCheckResource() resource.Resource {
	return &CheckResource{}
}

type CheckResourceModel struct {
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

func checkSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"check_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":           schema.StringAttribute{Required: true},
			"check_type":     schema.StringAttribute{Required: true},
			"period_seconds": schema.Int64Attribute{Required: true},
			"grace_seconds":  schema.Int64Attribute{Required: true},
			"url":            schema.StringAttribute{Optional: true},
			"expected_status_code": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(200),
			},
			"expected_string": schema.StringAttribute{Optional: true},
			"failure_threshold": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
			},
			"token": schema.StringAttribute{
				Computed: true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

func (r *CheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = checkSchema()
}

func (r *CheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func modelToCheck(m CheckResourceModel) Check {
	c := Check{
		Name:          m.Name.ValueString(),
		CheckType:     m.CheckType.ValueString(),
		PeriodSeconds: m.PeriodSeconds.ValueInt64(),
		GraceSeconds:  m.GraceSeconds.ValueInt64(),
	}
	if !m.URL.IsNull() && !m.URL.IsUnknown() {
		v := m.URL.ValueString()
		c.URL = &v
	}
	if !m.ExpectedStatusCode.IsNull() && !m.ExpectedStatusCode.IsUnknown() {
		v := m.ExpectedStatusCode.ValueInt64()
		c.ExpectedStatusCode = &v
	}
	if !m.ExpectedString.IsNull() && !m.ExpectedString.IsUnknown() {
		v := m.ExpectedString.ValueString()
		c.ExpectedString = &v
	}
	if !m.FailureThreshold.IsNull() && !m.FailureThreshold.IsUnknown() {
		v := m.FailureThreshold.ValueInt64()
		c.FailureThreshold = &v
	}
	return c
}

func checkToModel(teamID string, c *Check, m *CheckResourceModel) {
	m.TeamID = types.StringValue(teamID)
	m.CheckID = types.StringValue(c.CheckID)
	m.Name = types.StringValue(c.Name)
	m.CheckType = types.StringValue(c.CheckType)
	m.PeriodSeconds = types.Int64Value(c.PeriodSeconds)
	m.GraceSeconds = types.Int64Value(c.GraceSeconds)
	m.Token = types.StringValue(c.Token)

	if c.URL != nil {
		m.URL = types.StringValue(*c.URL)
	} else {
		m.URL = types.StringNull()
	}
	if c.ExpectedStatusCode != nil {
		m.ExpectedStatusCode = types.Int64Value(*c.ExpectedStatusCode)
	} else {
		m.ExpectedStatusCode = types.Int64Value(200)
	}
	if c.ExpectedString != nil {
		m.ExpectedString = types.StringValue(*c.ExpectedString)
	} else {
		m.ExpectedString = types.StringNull()
	}
	if c.FailureThreshold != nil {
		m.FailureThreshold = types.Int64Value(*c.FailureThreshold)
	} else {
		m.FailureThreshold = types.Int64Value(1)
	}
}

func (r *CheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.CreateCheck(ctx, plan.TeamID.ValueString(), modelToCheck(plan))
	if err != nil {
		resp.Diagnostics.AddError("Create failed", err.Error())
		return
	}
	checkToModel(plan.TeamID.ValueString(), check, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.GetCheck(ctx, state.TeamID.ValueString(), state.CheckID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read failed", err.Error())
		return
	}
	if check == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	checkToModel(state.TeamID.ValueString(), check, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.UpdateCheck(ctx, plan.TeamID.ValueString(), state.CheckID.ValueString(), modelToCheck(plan))
	if err != nil {
		resp.Diagnostics.AddError("Update failed", err.Error())
		return
	}
	checkToModel(plan.TeamID.ValueString(), check, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCheck(ctx, state.TeamID.ValueString(), state.CheckID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete failed", err.Error())
	}
}
