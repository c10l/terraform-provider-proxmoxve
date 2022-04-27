package proxmoxve

import (
	"context"
	"reflect"

	"github.com/c10l/proxmoxve-client-go/api2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type Version struct {
	ID      types.String `tfsdk:"id"`
	Release types.String `tfsdk:"release"`
	RepoID  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
}

type Pool struct {
	ID             types.String `tfsdk:"id"`
	PoolID         types.String `tfsdk:"pool_id"`
	Comment        types.String `tfsdk:"comment"`
	StorageMembers types.List   `tfsdk:"storage_members"`
	// Qemu    []PoolMemberQemu    `tfsdk:"qemu"`
	// LXC     []PoolMemberLXC     `tfsdk:"lxc"`
}

type PoolMemberStorage struct {
	Content    types.List   `tfsdk:"content"`
	Disk       types.Int64  `tfsdk:"disk"`
	ID         types.String `tfsdk:"id"`
	MaxDisk    types.Int64  `tfsdk:"max_disk"`
	Node       types.String `tfsdk:"node"`
	PluginType types.String `tfsdk:"plugin_type"`
	Shared     types.Int64  `tfsdk:"shared"`
	Status     types.String `tfsdk:"status"`
	Storage    types.String `tfsdk:"storage"`
}

func (s PoolMemberStorage) Type(ctx context.Context) attr.Type {
	return types.ObjectType{}
}

func (s PoolMemberStorage) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	contentList := []tftypes.Value{}
	for _, i := range s.Content.Elems {
		tfValue, err := i.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.Value{}, err
		}
		contentList = append(contentList, tfValue)
	}
	pmsType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"content":     tftypes.List{ElementType: tftypes.String},
			"disk":        tftypes.Number,
			"id":          tftypes.String,
			"max_disk":    tftypes.Number,
			"node":        tftypes.String,
			"plugin_type": tftypes.String,
			"shared":      tftypes.Number,
			"status":      tftypes.String,
			"storage":     tftypes.String,
		},
	}
	pmsValue := map[string]tftypes.Value{
		"content":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, contentList),
		"disk":        tftypes.NewValue(tftypes.Number, s.Disk.Value),
		"id":          tftypes.NewValue(tftypes.String, s.ID.Value),
		"max_disk":    tftypes.NewValue(tftypes.Number, s.MaxDisk.Value),
		"node":        tftypes.NewValue(tftypes.String, s.Node.Value),
		"plugin_type": tftypes.NewValue(tftypes.String, s.PluginType.Value),
		"shared":      tftypes.NewValue(tftypes.Number, s.Shared.Value),
		"status":      tftypes.NewValue(tftypes.String, s.Status.Value),
		"storage":     tftypes.NewValue(tftypes.String, s.Storage.Value),
	}
	return tftypes.NewValue(pmsType, pmsValue), nil
}

func (s PoolMemberStorage) Equal(v attr.Value) bool {
	return reflect.DeepEqual(s, v)
}

// type PoolMemberVM struct{}
// type PoolMemberQemu PoolMemberVM
// type PoolMemberLXC PoolMemberVM

type Storage struct {
	Storage types.String            `tfsdk:"id"`
	Content api2.StorageContentList `tfsdk:"content"`
	Digest  types.String            `tfsdk:"digest"`
	Path    types.String            `tfsdk:"path"`
	Type    api2.StorageType        `tfsdk:"type"`
}
