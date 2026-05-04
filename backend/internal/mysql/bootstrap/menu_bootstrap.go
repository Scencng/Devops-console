package bootstrap

import (
	"time"

	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type menuSeed struct {
	ID        uint64
	ParentID  uint64
	Name      string
	Type      int8
	Path      *string
	Component *string
	Icon      *string
	Perm      *string
	Sort      int
	Visible   int8
	Status    int8
}

func strPtr(v string) *string {
	return &v
}

var mysqlMenuSeeds = []menuSeed{
	{ID: 900, ParentID: 0, Name: "MySQL", Type: 1, Icon: strPtr("Coin"), Sort: 35, Visible: 1, Status: 1},
	{ID: 901, ParentID: 900, Name: "工作台（连接管理）", Type: 2, Path: strPtr("/mysql/workbench"), Component: strPtr("mysql/ConnectionManagement"), Icon: strPtr("Connection"), Perm: strPtr("mysql:connection:view"), Sort: 10, Visible: 1, Status: 1},
	{ID: 902, ParentID: 900, Name: "数据库管理", Type: 2, Path: strPtr("/mysql/databases"), Component: strPtr("mysql/DatabaseManagement"), Icon: strPtr("Coin"), Perm: strPtr("mysql:metadata:view"), Sort: 20, Visible: 1, Status: 1},
	{ID: 903, ParentID: 900, Name: "SQL 查询", Type: 2, Path: strPtr("/mysql/query"), Component: strPtr("mysql/SqlQuery"), Icon: strPtr("Document"), Perm: strPtr("mysql:query:view"), Sort: 30, Visible: 1, Status: 1},
	{ID: 904, ParentID: 900, Name: "数据管理", Type: 2, Path: strPtr("/mysql/data"), Component: strPtr("mysql/DataManagement"), Icon: strPtr("DataBoard"), Perm: strPtr("mysql:data:view"), Sort: 40, Visible: 1, Status: 1},
	{ID: 905, ParentID: 900, Name: "用户与权限", Type: 2, Path: strPtr("/mysql/security"), Component: strPtr("mysql/UserPermissionManagement"), Icon: strPtr("Lock"), Perm: strPtr("mysql:security:view"), Sort: 50, Visible: 1, Status: 1},
	{ID: 906, ParentID: 900, Name: "备份与恢复", Type: 2, Path: strPtr("/mysql/backup"), Component: strPtr("mysql/BackupManagement"), Icon: strPtr("FolderOpened"), Perm: strPtr("mysql:backup:view"), Sort: 60, Visible: 1, Status: 1},

	{ID: 9100, ParentID: 901, Name: "打开连接工作台", Type: 3, Perm: strPtr("mysql:connection:open"), Sort: 10, Visible: 1, Status: 1},
	{ID: 9200, ParentID: 902, Name: "创建数据库", Type: 3, Perm: strPtr("mysql:database:create"), Sort: 10, Visible: 1, Status: 1},
	{ID: 9201, ParentID: 902, Name: "重命名数据库", Type: 3, Perm: strPtr("mysql:database:rename"), Sort: 20, Visible: 1, Status: 1},
	{ID: 9202, ParentID: 902, Name: "删除数据库", Type: 3, Perm: strPtr("mysql:database:delete"), Sort: 30, Visible: 1, Status: 1},
	{ID: 9203, ParentID: 902, Name: "创建数据表", Type: 3, Perm: strPtr("mysql:table:create"), Sort: 40, Visible: 1, Status: 1},
	{ID: 9204, ParentID: 902, Name: "重命名数据表", Type: 3, Perm: strPtr("mysql:table:rename"), Sort: 50, Visible: 1, Status: 1},
	{ID: 9205, ParentID: 902, Name: "删除数据表", Type: 3, Perm: strPtr("mysql:table:delete"), Sort: 60, Visible: 1, Status: 1},

	{ID: 9300, ParentID: 903, Name: "执行 SQL", Type: 3, Perm: strPtr("mysql:query:execute"), Sort: 10, Visible: 1, Status: 1},

	{ID: 9400, ParentID: 904, Name: "新增数据", Type: 3, Perm: strPtr("mysql:data:create"), Sort: 10, Visible: 1, Status: 1},
	{ID: 9401, ParentID: 904, Name: "编辑数据", Type: 3, Perm: strPtr("mysql:data:update"), Sort: 20, Visible: 1, Status: 1},
	{ID: 9402, ParentID: 904, Name: "删除数据", Type: 3, Perm: strPtr("mysql:data:delete"), Sort: 30, Visible: 1, Status: 1},
	{ID: 9403, ParentID: 904, Name: "保存数据变更", Type: 3, Perm: strPtr("mysql:data:save"), Sort: 40, Visible: 1, Status: 1},
	{ID: 9404, ParentID: 904, Name: "导入数据", Type: 3, Perm: strPtr("mysql:data:import"), Sort: 50, Visible: 1, Status: 1},
	{ID: 9405, ParentID: 904, Name: "导出数据", Type: 3, Perm: strPtr("mysql:data:export"), Sort: 60, Visible: 1, Status: 1},

	{ID: 9500, ParentID: 905, Name: "管理用户与权限", Type: 3, Perm: strPtr("mysql:security:manage"), Sort: 10, Visible: 1, Status: 1},

	{ID: 9600, ParentID: 906, Name: "创建备份", Type: 3, Perm: strPtr("mysql:backup:create"), Sort: 10, Visible: 1, Status: 1},
	{ID: 9601, ParentID: 906, Name: "恢复备份", Type: 3, Perm: strPtr("mysql:backup:restore"), Sort: 20, Visible: 1, Status: 1},
	{ID: 9602, ParentID: 906, Name: "下载备份", Type: 3, Perm: strPtr("mysql:backup:download"), Sort: 30, Visible: 1, Status: 1},
	{ID: 9603, ParentID: 906, Name: "重命名备份", Type: 3, Perm: strPtr("mysql:backup:rename"), Sort: 40, Visible: 1, Status: 1},
	{ID: 9604, ParentID: 906, Name: "删除备份", Type: 3, Perm: strPtr("mysql:backup:delete"), Sort: 50, Visible: 1, Status: 1},
	{ID: 9605, ParentID: 906, Name: "创建备份计划", Type: 3, Perm: strPtr("mysql:backup:schedule:create"), Sort: 60, Visible: 1, Status: 1},
	{ID: 9606, ParentID: 906, Name: "删除备份计划", Type: 3, Perm: strPtr("mysql:backup:schedule:delete"), Sort: 70, Visible: 1, Status: 1},
}

var legacyMySQLMenuIDs = []uint64{7200, 7201, 7202, 7203, 7204, 7205, 7206}

func EnsureMySQLMenus(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	now := time.Now()

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SysMenu{}).
			Where("id IN ?", legacyMySQLMenuIDs).
			Updates(map[string]interface{}{
				"status":     0,
				"visible":    0,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		for _, item := range mysqlMenuSeeds {
			menu := model.SysMenu{
				ID:        item.ID,
				ParentID:  item.ParentID,
				Name:      item.Name,
				Type:      item.Type,
				Path:      item.Path,
				Component: item.Component,
				Icon:      item.Icon,
				Perm:      item.Perm,
				Sort:      item.Sort,
				Visible:   item.Visible,
				Status:    item.Status,
				UpdatedAt: &now,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"parent_id", "name", "type", "path", "component", "icon", "perm", "sort", "visible", "status", "updated_at"}),
			}).Create(&menu).Error; err != nil {
				return err
			}
		}

		var adminRoles []model.SysRole
		if err := tx.Where("code = ? AND status = 1", "admin").Find(&adminRoles).Error; err != nil {
			return err
		}

		for _, role := range adminRoles {
			for _, item := range mysqlMenuSeeds {
				relation := model.SysRoleMenu{
					RoleID: role.ID,
					MenuID: item.ID,
				}
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&relation).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
