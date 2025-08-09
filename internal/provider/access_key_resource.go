package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/henrywhitaker3/terraform-provider-grarage/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccessKeyResource{}
var _ resource.ResourceWithImportState = &AccessKeyResource{}

func NewAccessKeyResource() resource.Resource {
	return &AccessKeyResource{}
}

// AccessKeyResource defines the resource implementation.
type AccessKeyResource struct {
	client *client.Client
}

// AccessKeyResourceModel describes the resource data model.
type AccessKeyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Expiration      types.String `tfsdk:"expiration"`
	NeverExpires    types.Bool   `tfsdk:"never_expires"`
}

func (r *AccessKeyResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_access_key"
}

func (r *AccessKeyResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Access key resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The id of the key",
				Computed:            true,
				Sensitive:           true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the access key",
				Required:            true,
			},
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: "The access key id",
				Computed:            true,
				Sensitive:           true,
			},
			"secret_access_key": schema.StringAttribute{
				MarkdownDescription: "The secret access key",
				Computed:            true,
				Sensitive:           true,
			},
			"expiration": schema.StringAttribute{
				MarkdownDescription: "The time in RFC 3339 format that the key should expire",
				Optional:            true,
				Computed:            true,
			},
			"never_expires": schema.BoolAttribute{
				MarkdownDescription: "Whether the key should expire or not",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *AccessKeyResource) Configure(
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

func (r *AccessKeyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data AccessKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.NeverExpires.ValueBool() && data.Expiration.ValueString() != "" {
		resp.Diagnostics.AddError(
			"invalid input",
			"cannot set never_expires and expiration together",
		)
		return
	}

	key, err := r.client.CreateAccessKey(ctx, client.CreateKeyRequest{
		Name:         data.Name.ValueString(),
		Expiration:   data.Expiration.ValueString(),
		NeverExpires: data.NeverExpires.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("could not create key", err.Error())
		return
	}

	mapKeyToData(&data, key)

	tflog.Trace(ctx, "created a bucket")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessKeyResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data AccessKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetAccessKey(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("could not get key", err.Error())
		return
	}

	mapKeyToData(&data, key)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessKeyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data AccessKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.UpdateAccessKey(
		ctx,
		data.ID.ValueString(),
		client.CreateKeyRequest{
			Expiration:   data.Expiration.ValueString(),
			NeverExpires: data.NeverExpires.ValueBool(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("could not update key", err.Error())
		return
	}

	mapKeyToData(&data, key)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapKeyToData(data *AccessKeyResourceModel, key *client.AccessKey) {
	data.ID = types.StringValue(key.AccessKeyID)
	data.Name = types.StringValue(key.Name)
	data.AccessKeyID = types.StringValue(key.AccessKeyID)
	data.Expiration = types.StringNull()
	if key.Expiration == nil {
		data.NeverExpires = types.BoolValue(true)
	} else {

		data.Expiration = types.StringValue(*key.Expiration)
	}
	data.SecretAccessKey = types.StringNull()
	if key.SecretAccessKey != nil {
		data.SecretAccessKey = types.StringValue(*key.SecretAccessKey)
	}
}

func (r *AccessKeyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data AccessKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAccessKey(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("could not delete key", err.Error())
		return
	}
}

func (r *AccessKeyResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
