package provider

import (
	"context"
	"fmt"
	"io"
	"time"

	garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccessKeyResource{}
var _ resource.ResourceWithImportState = &AccessKeyResource{}

func NewAccessKeyResource() resource.Resource {
	return &AccessKeyResource{}
}

// AccessKeyResource defines the resource implementation.
type AccessKeyResource struct {
	client *garage.APIClient
	ctx    context.Context
}

// AccessKeyResourceModel describes the resource data model.
type AccessKeyResourceModel struct {
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
			},
			"never_expires": schema.BoolAttribute{
				MarkdownDescription: "Whether the key should expire or not",
				Optional:            true,
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
	r.ctx = setup.ctx
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

	body := garage.UpdateKeyRequestBody{
		Name:         *garage.NewNullableString(data.Name.ValueStringPointer()),
		NeverExpires: data.NeverExpires.ValueBoolPointer(),
	}

	if exp := data.Expiration.ValueString(); exp != "" {
		et, err := time.Parse(time.RFC3339, exp)
		if err != nil {
			resp.Diagnostics.AddError("invalid expiration value", err.Error())
			return
		}
		body.Expiration = *garage.NewNullableTime(&et)
	}

	key, _, err := r.client.AccessKeyAPI.CreateKey(r.ctx).
		Body(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("could not create access key", err.Error())
		return
	}

	data.Name = types.StringValue(key.Name)
	data.AccessKeyID = types.StringValue(key.AccessKeyId)
	data.Expiration = types.StringValue(key.Expiration.Get().String())
	data.NeverExpires = types.BoolValue(!key.HasExpiration())
	if sec := key.SecretAccessKey.Get(); sec != nil {
		data.SecretAccessKey = types.StringValue(*sec)
	}

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

	bucket, _, err := r.client.BucketAPI.GetBucketInfo(r.ctx).Id(data.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"could not get bucket",
			fmt.Sprintf("got error from client: %s", err),
		)
	}

	data.ID = types.StringValue(bucket.Id)
	data.Name = types.StringValue(bucket.GlobalAliases[0])

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

	bucket, _, err := r.client.BucketAPI.UpdateBucket(r.ctx, data.ID.ValueString()).
		UpdateBucketRequestBody(garage.UpdateBucketRequestBody{}).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError("could not update bucket", err.Error())
		return
	}

	data.ID = types.StringValue(bucket.Id)
	data.Name = types.StringValue(bucket.GlobalAliases[0])

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	response, err := r.client.BucketAPI.DeleteBucket(r.ctx, data.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError("could not delete bucket", err.Error())
		return
	}
	if response.StatusCode > 299 {
		body, _ := io.ReadAll(response.Body)
		resp.Diagnostics.AddError(
			"could not delete bucket",
			fmt.Sprintf("got status code %d: %s", response.StatusCode, string(body)),
		)
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
