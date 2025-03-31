package entimport_test

import (
	"testing"

	"ariga.io/atlas/sql/schema"
	"entgo.io/contrib/schemast"
	"github.com/powerfulyang/entimport/internal/entimport"
	"github.com/stretchr/testify/require"
)

// 测试 upsertOneToX 函数是否能正确处理同一表多次引用的情况
func TestMultipleReferencesToSameTable(t *testing.T) {
	// 创建模拟表结构，其中 videos 表有多个外键引用 files 表
	filesTable := &schema.Table{
		Name: "files",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "text"},
					Raw:  "text",
					Null: false,
				},
			},
		},
	}
	filesTable.PrimaryKey = &schema.Index{
		Name:   "files_pkey",
		Unique: true,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				C:     filesTable.Columns[0],
			},
		},
	}

	videosTable := &schema.Table{
		Name: "videos",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "text"},
					Raw:  "text",
					Null: false,
				},
			},
			{
				Name: "posterId",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "text"},
					Raw:  "text",
					Null: true,
				},
			},
			{
				Name: "fileId",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "text"},
					Raw:  "text",
					Null: false,
				},
			},
			{
				Name: "thumbnail_id",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "text"},
					Raw:  "text",
					Null: true,
				},
			},
		},
	}
	videosTable.PrimaryKey = &schema.Index{
		Name:   "videos_pkey",
		Unique: true,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				C:     videosTable.Columns[0],
			},
		},
	}
	videosTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol:     "videos_posterId_fkey",
			Table:      videosTable,
			Columns:    []*schema.Column{videosTable.Columns[1]},
			RefTable:   filesTable,
			RefColumns: []*schema.Column{filesTable.Columns[0]},
			OnUpdate:   "NO ACTION",
			OnDelete:   "SET NULL",
		},
		{
			Symbol:     "videos_fileId_fkey",
			Table:      videosTable,
			Columns:    []*schema.Column{videosTable.Columns[2]},
			RefTable:   filesTable,
			RefColumns: []*schema.Column{filesTable.Columns[0]},
			OnUpdate:   "NO ACTION",
			OnDelete:   "CASCADE",
		},
		{
			Symbol:     "videos_thumbnail_id_fkey",
			Table:      videosTable,
			Columns:    []*schema.Column{videosTable.Columns[3]},
			RefTable:   filesTable,
			RefColumns: []*schema.Column{filesTable.Columns[0]},
			OnUpdate:   "NO ACTION",
			OnDelete:   "SET NULL",
		},
	}

	// 创建 mutations 映射
	mutations := make(map[string]schemast.Mutator)
	mutations["files"] = &schemast.UpsertSchema{Name: "File"}
	mutations["videos"] = &schemast.UpsertSchema{Name: "Video"}

	// 调用 upsertOneToX
	t.Logf("videos 表有 %d 个外键", len(videosTable.ForeignKeys))
	for i, fk := range videosTable.ForeignKeys {
		t.Logf("外键 %d: %s -> %s.%s", i, fk.Columns[0].Name, fk.RefTable.Name, fk.RefColumns[0].Name)
	}

	entimport.TestableUpsertOneToX(mutations, videosTable)

	// 获取 Video schema
	videoSchema, ok := mutations["videos"].(*schemast.UpsertSchema)
	filesSchema, ok := mutations["files"].(*schemast.UpsertSchema)
	require.True(t, ok)

	// 详细的边信息
	t.Logf("\n videoSchema:")
	for i, e := range videoSchema.Edges {
		desc := e.Descriptor()
		t.Logf("边 %d 详情:", i)
		t.Logf("  - 名称(Name): %s", desc.Name)
		t.Logf("  - 类型(Type): %s", desc.Type)
		t.Logf("  - 引用名(RefName): %s", desc.RefName)
		t.Logf("  - 字段名(Field): %s", desc.Field)
		t.Logf("  - 是否唯一(Unique): %v", desc.Unique)
		t.Logf("  - 是否反向(Inverse): %v", desc.Inverse)
	}

	// files
	t.Logf("\n filesSchema:")
	for i, e := range filesSchema.Edges {
		desc := e.Descriptor()
		t.Logf("边 %d 详情:", i)
		t.Logf("  - 名称(Name): %s", desc.Name)
		t.Logf("  - 类型(Type): %s", desc.Type)
		t.Logf("  - 引用名(RefName): %s", desc.RefName)
		t.Logf("  - 字段名(Field): %s", desc.Field)
		t.Logf("  - 是否唯一(Unique): %v", desc.Unique)
		t.Logf("  - 是否反向(Inverse): %v", desc.Inverse)
	}

	// 检查边是否使用了正确的字段名
	edgeFields := make(map[string]bool)
	for _, e := range videoSchema.Edges {
		desc := e.Descriptor()
		if desc.Type == "File" && desc.Field != "" {
			edgeFields[desc.Name] = true
		}
	}

	// 验证是否使用了正确的字段名
	require.True(t, edgeFields["poster"], "应该有Field为'poster'的边")
	require.True(t, edgeFields["file"], "应该有Field为'file'的边")
	require.True(t, edgeFields["thumbnail"], "应该有Field为'thumbnail'的边")
}
