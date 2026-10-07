// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/reiinoldo/terraform-provider-aidbox-license/internal/aidbox"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LicenseResource{}
var _ resource.ResourceWithImportState = &LicenseResource{}

func NewLicenseResource() resource.Resource {
	return &LicenseResource{}
}

// LicenseResource defines the resource implementation.
type LicenseResource struct {
	client Client
}

// LicenseResourceModel describes the resource data model.
type LicenseResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Product         types.String `tfsdk:"product"`
	Type            types.String `tfsdk:"type"`
	Expiration      types.String `tfsdk:"expiration"`
	Status          types.String `tfsdk:"status"`
	MaxInstances    types.Int64  `tfsdk:"max_instances"`
	CreatorID       types.String `tfsdk:"creator_id"`
	ProjectID       types.String `tfsdk:"project_id"`
	Offline         types.Bool   `tfsdk:"offline"`
	Created         types.String `tfsdk:"created"`
	MetaLastUpdated types.String `tfsdk:"meta_last_updated"`
	MetaCreatedAt   types.String `tfsdk:"meta_created_at"`
	MetaVersionID   types.String `tfsdk:"meta_version_id"`
	Issuer          types.String `tfsdk:"issuer"`
	InfoHosting     types.String `tfsdk:"info_hosting"`
	JWT             types.String `tfsdk:"jwt"`
}

func (r *LicenseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (r *LicenseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Aidbox license. Changes to the name, product, or type replace the license; destroying this resource asks Aidbox to remove it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"product": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("aidbox"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"max_instances": schema.Int64Attribute{
				Computed: true,
			},
			"creator_id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Computed: true,
			},
			"offline": schema.BoolAttribute{
				Computed: true,
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
			"meta_last_updated": schema.StringAttribute{
				Computed: true,
			},
			"meta_created_at": schema.StringAttribute{
				Computed: true,
			},
			"meta_version_id": schema.StringAttribute{
				Computed: true,
			},
			"issuer": schema.StringAttribute{
				Computed: true,
			},
			"info_hosting": schema.StringAttribute{
				Computed: true,
			},
			"jwt": schema.StringAttribute{
				MarkdownDescription: "The signed Aidbox license token. Marked sensitive, but still stored in Terraform or OpenTofu state.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *LicenseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = data.Client
}

func (r *LicenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model LicenseResourceModel
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.CreateLicense(ctx, model.Name.ValueString(), model.Product.ValueString(), model.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Call Failed", err.Error())
		return
	}

	mapModelFromAPIResponse(&model, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *LicenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model LicenseResourceModel

	// Read Terraform state data into the model
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure the ID attribute is present
	if model.ID.IsUnknown() || model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"No ID Present",
			"An ID must be present to read the License",
		)
		return
	}

	// Use the client to fetch the license data from the API
	apiResp, err := r.client.GetLicense(ctx, model.ID.ValueString())
	if err != nil {
		if errors.Is(err, aidbox.ErrLicenseNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to Fetch License", fmt.Sprintf("Unable to fetch license: %s", err))
		return
	}

	// Map the API response back to the Terraform model
	mapModelFromAPIResponse(&model, apiResp)

	// Save the updated model back into the Terraform state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *LicenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"License Updates Require Replacement",
		"All configurable license attributes require replacement. Please report this error if it occurs after Terraform or OpenTofu planned a replacement.",
	)
}

func (r *LicenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model LicenseResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure the ID attribute is present
	if model.ID.IsNull() || model.ID.IsUnknown() || model.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"No ID Found",
			"Cannot delete the License without an ID.",
		)
		return
	}

	// Call the DeleteLicense method from the AidboxHTTPClient with the ID from the model
	err := r.client.DeleteLicense(ctx, model.ID.ValueString())
	if err != nil && !errors.Is(err, aidbox.ErrLicenseNotFound) {
		resp.Diagnostics.AddError(
			"Failed to Delete License",
			fmt.Sprintf("Error while trying to delete the License with ID %s: %s", model.ID.ValueString(), err.Error()),
		)
		return
	}

	// Successfully deleted the License, indicate this by marking the resource as removed
	resp.State.RemoveResource(ctx)
}

func (r *LicenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapModelFromAPIResponse(model *LicenseResourceModel, apiResp aidbox.LicenseResponse) {
	model.ID = basetypes.NewStringValue(apiResp.License.ID)
	model.Name = basetypes.NewStringValue(apiResp.License.Name)
	model.Product = basetypes.NewStringValue(apiResp.License.Product)
	model.Type = basetypes.NewStringValue(apiResp.License.Type)
	model.Expiration = basetypes.NewStringValue(apiResp.License.Expiration)
	model.Status = basetypes.NewStringValue(apiResp.License.Status)
	model.MaxInstances = basetypes.NewInt64Value(int64(apiResp.License.MaxInstances))
	model.CreatorID = basetypes.NewStringValue(apiResp.License.Creator.ID)
	model.ProjectID = basetypes.NewStringValue(apiResp.License.Project.ID)
	model.Offline = basetypes.NewBoolValue(apiResp.License.Offline)
	model.Created = basetypes.NewStringValue(apiResp.License.Created)
	model.MetaLastUpdated = basetypes.NewStringValue(apiResp.License.Meta.LastUpdated)
	model.MetaCreatedAt = basetypes.NewStringValue(apiResp.License.Meta.CreatedAt)
	model.MetaVersionID = basetypes.NewStringValue(apiResp.License.Meta.VersionID)
	model.Issuer = basetypes.NewStringValue(apiResp.License.Issuer)
	model.InfoHosting = basetypes.NewStringValue(apiResp.License.Info.Hosting)
	model.JWT = basetypes.NewStringValue(apiResp.JWT)
}
