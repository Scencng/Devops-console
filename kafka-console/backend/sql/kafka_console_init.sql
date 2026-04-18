SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `system_user_token`;
DROP TABLE IF EXISTS `sys_user_position`;
DROP TABLE IF EXISTS `sys_user_role`;
DROP TABLE IF EXISTS `sys_role_menu`;
DROP TABLE IF EXISTS `sys_menu`;
DROP TABLE IF EXISTS `sys_role`;
DROP TABLE IF EXISTS `sys_position`;
DROP TABLE IF EXISTS `sys_department`;
DROP TABLE IF EXISTS `system_users`;

CREATE TABLE `sys_department` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT 0,
  `name` varchar(100) NOT NULL,
  `sort` int NOT NULL DEFAULT 0,
  `leader` varchar(50) DEFAULT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT 1,
  `remark` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_sys_department_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sys_position` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `code` varchar(100) NOT NULL,
  `sort` int NOT NULL DEFAULT 0,
  `status` tinyint unsigned NOT NULL DEFAULT 1,
  `remark` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_position_code` (`code`),
  KEY `idx_sys_position_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sys_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `code` varchar(100) NOT NULL,
  `sort` int NOT NULL DEFAULT 0,
  `status` tinyint unsigned NOT NULL DEFAULT 1,
  `remark` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_role_code` (`code`),
  KEY `idx_sys_role_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sys_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT 0,
  `name` varchar(100) NOT NULL,
  `type` tinyint NOT NULL DEFAULT 1,
  `path` varchar(200) DEFAULT NULL,
  `component` varchar(200) DEFAULT NULL,
  `icon` varchar(50) DEFAULT NULL,
  `perm` varchar(200) DEFAULT NULL,
  `sort` int NOT NULL DEFAULT 0,
  `visible` tinyint NOT NULL DEFAULT 1,
  `status` tinyint NOT NULL DEFAULT 1,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_sys_menu_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sys_role_menu` (
  `role_id` bigint unsigned NOT NULL,
  `menu_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`,`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `system_users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(191) NOT NULL,
  `password` varchar(191) NOT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT 1,
  `nickname` varchar(191) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `avatar` varchar(500) DEFAULT NULL,
  `dept_id` bigint unsigned DEFAULT NULL,
  `remark` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_system_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sys_user_role` (
  `user_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sys_user_position` (
  `user_id` bigint unsigned NOT NULL,
  `position_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`user_id`,`position_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `system_user_token` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `refresh_token` varchar(255) NOT NULL,
  `expires_at` datetime NOT NULL,
  `last_login_ip` varchar(45) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `access_token` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `sys_department` VALUES (1,0,'默认部门',0,NULL,NULL,NULL,1,'默认部门',NOW(3),NOW(3),NULL);
INSERT INTO `sys_position` VALUES (1,'管理员','admin',0,1,'默认管理员岗位',NOW(3),NOW(3),NULL);
INSERT INTO `sys_role` VALUES (1,'超级管理员','admin',0,1,'系统管理员',NOW(3),NOW(3),NULL);
INSERT INTO `system_users` VALUES (1,'admin','240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9',1,'管理员',NULL,NULL,NULL,1,'默认管理员',NOW(3),NOW(3),NULL);
INSERT INTO `sys_user_role` VALUES (1,1);
INSERT INTO `sys_user_position` VALUES (1,1);

INSERT INTO `sys_menu` VALUES
(500,0,'系统管理',1,'','', 'Setting',NULL,90,1,1,NOW(3),NOW(3),NULL),
(501,500,'用户管理',2,'/system/users','system/UserManagement','User',NULL,10,1,1,NOW(3),NOW(3),NULL),
(502,500,'部门管理',2,'/system/depts','system/DeptManagement','OfficeBuilding',NULL,20,1,1,NOW(3),NOW(3),NULL),
(503,500,'岗位管理',2,'/system/positions','system/PositionManagement','Briefcase',NULL,30,1,1,NOW(3),NOW(3),NULL),
(504,500,'角色管理',2,'/system/roles','system/RoleManagement','UserFilled',NULL,40,1,1,NOW(3),NOW(3),NULL),
(505,500,'菜单管理',2,'/system/menus','system/MenuManagement','Menu',NULL,50,1,1,NOW(3),NOW(3),NULL);

INSERT INTO `sys_role_menu` VALUES (1,500),(1,501),(1,502),(1,503),(1,504),(1,505);

SET FOREIGN_KEY_CHECKS = 1;
