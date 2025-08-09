package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/henrywhitaker3/terraform-provider-garage/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PermissionResource{}
var _ resource.ResourceWithImportState = &PermissionResource{}

func NewPermissionResource() resource.Resource {
	return &PermissionResource{}
}

// PermissionResource defines the resource implementation.
type PermissionResource struct {
	client *client.Client
}

// PermissionResourceModel describes the resource data model.
type PermissionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	AccessKeyID types.String `tfsdk:"access_key_id"`
	BucketID    types.String `tfsdk:"bucket_id"`
	Owner       types.Bool   `tfsdk:"owner"`
	Read        types.Bool   `tfsdk:"read"`
	Write       types.Bool   `tfsdk:"write"`
}

func (r *PermissionResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *PermissionResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Access key resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The id of the permission in format {accessKeyId}:{bucketId}",
				Computed:            true,
				Sensitive:           true,
			},
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: "The access key id",
				Required:            true,
			},
			"bucket_id": schema.StringAttribute{
				MarkdownDescription: "The bucket id",
				Required:            true,
			},
			"owner": schema.BoolAttribute{
				MarkdownDescription: "Whether the key is the owner of the bucket",
				Optional:            true,
				Computed:            true,
			},
			"read": schema.BoolAttribute{
				MarkdownDescription: "Whether the key can read from the bucket",
				Optional:            true,
				Computed:            true,
			},
			"write": schema.BoolAttribute{
				MarkdownDescription: "Whether the key can write to the bucket",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *PermissionResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	setup, ok := req.ProviderData.(setupData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected setupData, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = setup.client
}

func (r *PermissionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data PermissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	perms, err := r.client.CreatePermission(ctx, client.CreatePermissionRequest{
		AccessKeyID: data.AccessKeyID.ValueString(),
		BucketID:    data.BucketID.ValueString(),
		Permissions: client.CreatePermissionsBlock{
			Owner: data.Owner.ValueBool(),
			Read:  data.Read.ValueBool(),
			Write: data.Write.ValueBool(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("could not create permissions", err.Error())
		return
	}

	mapPermsToData(&data, perms)

	tflog.Trace(ctx, "created a permission")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data PermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	spl := strings.Split(data.ID.ValueString(), ":")
	if len(spl) != 2 {
		resp.Diagnostics.AddError(
			"invalid permission id",
			fmt.Sprintf("needs id in format {keyId}:{bucketId}, got %s", data.ID.ValueString()),
		)
		return
	}

	perms, err := r.client.GetPermissions(ctx, spl[0], spl[1])
	if err != nil {
		resp.Diagnostics.AddError("could not get permissions", err.Error())
		return
	}

	mapPermsToData(&data, perms)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data PermissionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	perms, err := r.client.UpdatePermission(ctx, client.CreatePermissionRequest{
		AccessKeyID: data.AccessKeyID.ValueString(),
		BucketID:    data.BucketID.ValueString(),
		Permissions: client.CreatePermissionsBlock{
			Owner: data.Owner.ValueBool(),
			Read:  data.Read.ValueBool(),
			Write: data.Write.ValueBool(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("could not update permissions", err.Error())
		return
	}

	mapPermsToData(&data, perms)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapPermsToData(data *PermissionResourceModel, perms *client.Permission) {
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", perms.AccessKeyID, perms.BucketID))
	data.AccessKeyID = types.StringValue(perms.AccessKeyID)
	data.BucketID = types.StringValue(perms.BucketID)
	data.Owner = types.BoolValue(perms.Owner)
	data.Read = types.BoolValue(perms.Read)
	data.Write = types.BoolValue(perms.Write)
}

func (r *PermissionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data PermissionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *PermissionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
