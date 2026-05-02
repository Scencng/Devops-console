package configs

import (
	"devops-console-backend/internal/dal"
	"devops-console-backend/internal/dal/model"
	"errors"

	"gorm.io/gorm"
)

func AutoMigrateCustomTables() error {
	if Config == nil || !Config.Database.AutoMigrate || GORMDB == nil {
		return nil
	}
	if err := GORMDB.AutoMigrate(
		&dal.KafkaCluster{},
		&dal.KafkaAuditLog{},
	); err != nil {
		return err
	}
	if err := seedKafkaInstanceType(); err != nil {
		return err
	}
	return seedKafkaMenus()
}

func seedKafkaInstanceType() error {
	instanceType := dal.InstanceType{}
	err := GORMDB.Where("type_name = ?", "Kafka").First(&instanceType).Error
	if err == nil {
		if instanceType.Description == "" {
			return GORMDB.Model(&instanceType).Update("description", "Kafka 消息队列").Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return GORMDB.Create(&dal.InstanceType{
		TypeName:    "Kafka",
		Description: "Kafka 消息队列",
	}).Error
}

func seedKafkaMenus() error {
	if err := cleanupLegacyKafkaDashboardMenu(); err != nil {
		return err
	}
	menus := []model.SysMenu{
		menuDir(7000, 0, "Kafka", "Connection", 20),
		menuPage(7002, 7000, "集群管理", "/kafka/clusters", "kafka/ClusterManagement", "Connection", 20),
		menuPage(7010, 7000, "自动发现", "/kafka/discovery", "kafka/DiscoveryCenter", "Search", 25),
		menuPage(7003, 7000, "Topic 管理", "/kafka/topics", "kafka/TopicManagement", "DocumentCopy", 30),
		menuPage(7004, 7000, "Broker 管理", "/kafka/brokers", "kafka/BrokerManagement", "Monitor", 40),
		menuPage(7005, 7000, "消费组管理", "/kafka/groups", "kafka/ConsumerGroupManagement", "Histogram", 50),
		menuPage(7006, 7000, "消息浏览", "/kafka/messages", "kafka/MessageBrowser", "Search", 60),
		menuPage(7008, 7000, "审计日志", "/kafka/audits", "kafka/AuditLog", "Document", 80),
		menuPage(7101, 7002, "新增 Kafka 集群", "", "", "", 10, "kafka:cluster:create"),
		menuPage(7102, 7002, "编辑 Kafka 集群", "", "", "", 20, "kafka:cluster:edit"),
		menuPage(7103, 7002, "删除 Kafka 集群", "", "", "", 30, "kafka:cluster:delete"),
		menuPage(7104, 7002, "测试 Kafka 集群", "", "", "", 40, "kafka:cluster:test"),
		menuPage(7105, 7003, "修改 Topic 配置", "", "", "", 10, "kafka:topic:config:update"),
		menuPage(7106, 7003, "删除 Topic", "", "", "", 20, "kafka:topic:delete"),
		menuPage(7112, 7004, "修改 Broker 动态配置", "", "", "", 10, "kafka:broker:config:update"),
		menuPage(7107, 7005, "重置消费组 Offset", "", "", "", 10, "kafka:group:offset:reset"),
		menuPage(7111, 7005, "删除消费组", "", "", "", 20, "kafka:group:delete"),
		menuPage(7108, 7003, "创建 Topic", "", "", "", 30, "kafka:topic:create"),
		menuPage(7109, 7003, "扩容 Topic 分区", "", "", "", 40, "kafka:topic:partitions:increase"),
		menuPage(7110, 7006, "发送消息", "", "", "", 10, "kafka:message:produce"),
	}

	for _, item := range menus {
		if err := GORMDB.Where("id = ?", item.ID).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}

	roleMenuIDs := []uint64{7000, 7002, 7003, 7004, 7005, 7006, 7008, 7010, 7101, 7102, 7103, 7104, 7105, 7106, 7107, 7108, 7109, 7110, 7111, 7112}
	for _, menuID := range roleMenuIDs {
		roleMenu := model.SysRoleMenu{RoleID: 1, MenuID: menuID}
		if err := GORMDB.Where("role_id = ? AND menu_id = ?", roleMenu.RoleID, roleMenu.MenuID).FirstOrCreate(&roleMenu).Error; err != nil {
			return err
		}
	}

	return nil
}

func cleanupLegacyKafkaDashboardMenu() error {
	if err := GORMDB.Where("menu_id = ?", 7001).Delete(&model.SysRoleMenu{}).Error; err != nil {
		return err
	}
	return GORMDB.Where("id = ?", 7001).Delete(&model.SysMenu{}).Error
}

func menuDir(id, parentID uint64, name, icon string, sort int) model.SysMenu {
	path := ""
	component := ""
	perm := ""
	return model.SysMenu{
		ID:        id,
		ParentID:  parentID,
		Name:      name,
		Type:      1,
		Path:      &path,
		Component: &component,
		Icon:      &icon,
		Perm:      &perm,
		Sort:      sort,
		Visible:   1,
		Status:    1,
	}
}

func menuPage(id, parentID uint64, name, pathValue, componentValue, icon string, sort int, perms ...string) model.SysMenu {
	perm := ""
	if len(perms) > 0 {
		perm = perms[0]
	}
	return model.SysMenu{
		ID:        id,
		ParentID:  parentID,
		Name:      name,
		Type:      menuType(pathValue, componentValue, perm),
		Path:      stringPtr(pathValue),
		Component: stringPtr(componentValue),
		Icon:      stringPtr(icon),
		Perm:      stringPtr(perm),
		Sort:      sort,
		Visible:   1,
		Status:    1,
	}
}

func menuType(pathValue, componentValue, perm string) int8 {
	if perm != "" && pathValue == "" && componentValue == "" {
		return 3
	}
	return 2
}

func stringPtr(value string) *string {
	return &value
}
