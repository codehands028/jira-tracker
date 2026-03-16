-- Jira Tracker 数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS jira_tracker DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE jira_tracker;

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `phone` varchar(20) NOT NULL,
  `name` varchar(50) NOT NULL,
  `role` varchar(20) NOT NULL COMMENT 'admin/test/dev',
  `status` tinyint DEFAULT '1' COMMENT '1-启用 0-禁用',
  `ticket_num` int DEFAULT '0' COMMENT '当前待处理工单数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_phone` (`phone`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 工单表
CREATE TABLE IF NOT EXISTS `tickets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `jira_key` varchar(50) NOT NULL,
  `jira_url` varchar(500) NOT NULL,
  `description` text,
  `type` varchar(50) DEFAULT NULL COMMENT '工单类型(bug/feature/task/improvement)',
  `priority` varchar(20) DEFAULT NULL COMMENT 'low/medium/high/critical',
  `status` varchar(20) NOT NULL COMMENT 'processing/retesting/closed/timeout',
  `current_user_id` bigint unsigned NOT NULL,
  `is_timeout` tinyint DEFAULT '0',
  `timeout_level` varchar(20) DEFAULT NULL COMMENT 'normal/severe',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_jira_key` (`jira_key`),
  KEY `idx_tickets_deleted_at` (`deleted_at`),
  KEY `idx_current_user_id` (`current_user_id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 工单流转记录表
CREATE TABLE IF NOT EXISTS `ticket_flows` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `ticket_id` bigint unsigned NOT NULL,
  `from_user_id` bigint unsigned NOT NULL,
  `to_user_id` bigint unsigned NOT NULL,
  `content` text,
  `process_time` bigint DEFAULT NULL COMMENT '处理时长(秒)',
  `is_timeout` tinyint DEFAULT '0',
  `timeout_level` varchar(20) DEFAULT NULL COMMENT 'normal/severe',
  PRIMARY KEY (`id`),
  KEY `idx_ticket_flows_deleted_at` (`deleted_at`),
  KEY `idx_ticket_id` (`ticket_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知表
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL,
  `type` varchar(50) NOT NULL COMMENT 'timeout/flow/close',
  `title` varchar(200) NOT NULL,
  `content` text,
  `is_read` tinyint DEFAULT '0',
  `ticket_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notifications_deleted_at` (`deleted_at`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_ticket_id` (`ticket_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 操作日志表
CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL,
  `module` varchar(50) NOT NULL,
  `action` varchar(50) NOT NULL,
  `description` text,
  `ip_address` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_deleted_at` (`deleted_at`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 超时规则表
CREATE TABLE IF NOT EXISTS `timeout_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) NOT NULL,
  `normal_limit` bigint NOT NULL COMMENT '普通超时时间(秒)',
  `severe_limit` bigint NOT NULL COMMENT '严重超时时间(秒)',
  `is_active` tinyint DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_timeout_rules_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认超时规则
INSERT INTO `timeout_rules` (`created_at`, `updated_at`, `name`, `normal_limit`, `severe_limit`, `is_active`) VALUES
(NOW(), NOW(), '默认规则', 86400, 172800, 1),
(NOW(), NOW(), '紧急规则', 14400, 28800, 0),
(NOW(), NOW(), '宽松规则', 172800, 259200, 0);

-- SLA规则表
CREATE TABLE IF NOT EXISTS `sla_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) NOT NULL COMMENT '规则名称',
  `ticket_type` varchar(50) DEFAULT NULL COMMENT '工单类型(bug/feature/task等)',
  `priority` varchar(20) DEFAULT NULL COMMENT '优先级(low/medium/high/critical)',
  `project` varchar(100) DEFAULT NULL COMMENT '项目标识',
  `normal_limit` bigint NOT NULL COMMENT '普通SLA时限(秒)',
  `severe_limit` bigint NOT NULL COMMENT '严重SLA时限(秒)',
  `is_active` tinyint DEFAULT '1' COMMENT '是否启用',
  `priority_order` int DEFAULT '0' COMMENT '匹配优先级，数值越大优先级越高',
  PRIMARY KEY (`id`),
  KEY `idx_sla_rules_deleted_at` (`deleted_at`),
  KEY `idx_priority` (`priority`),
  KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认SLA规则
INSERT INTO `sla_rules` (`created_at`, `updated_at`, `name`, `ticket_type`, `priority`, `project`, `normal_limit`, `severe_limit`, `is_active`, `priority_order`) VALUES
(NOW(), NOW(), '紧急工单SLA', '', 'critical', '', 14400, 28800, 1, 10),
(NOW(), NOW(), '高优先级SLA', '', 'high', '', 28800, 57600, 1, 10),
(NOW(), NOW(), '中优先级SLA', '', 'medium', '', 86400, 172800, 1, 10),
(NOW(), NOW(), '低优先级SLA', '', 'low', '', 172800, 259200, 1, 10),
(NOW(), NOW(), '默认SLA', '', '', '', 86400, 172800, 1, 0);

-- 批量操作日志表
CREATE TABLE IF NOT EXISTS `batch_operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL COMMENT '操作人ID',
  `operation_type` varchar(50) NOT NULL COMMENT '操作类型(flow/assign/close)',
  `ticket_ids` text NOT NULL COMMENT '涉及的工单ID列表(JSON)',
  `ticket_count` int NOT NULL COMMENT '涉及工单数量',
  `success_count` int NOT NULL COMMENT '成功数量',
  `detail` text COMMENT '操作详情(JSON)',
  `ip_address` varchar(50) DEFAULT NULL COMMENT '操作IP地址',
  PRIMARY KEY (`id`),
  KEY `idx_batch_operation_logs_deleted_at` (`deleted_at`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_operation_type` (`operation_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认管理员账号
INSERT INTO `users` (`created_at`, `updated_at`, `phone`, `name`, `role`, `status`, `ticket_num`) VALUES
(NOW(), NOW(), '13800138000', '系统管理员', 'admin', 1, 0);

-- 插入测试用户（可选）
INSERT INTO `users` (`created_at`, `updated_at`, `phone`, `name`, `role`, `status`, `ticket_num`) VALUES
(NOW(), NOW(), '13800138001', '张三', 'test', 1, 0),
(NOW(), NOW(), '13800138002', '李四', 'dev', 1, 0),
(NOW(), NOW(), '13800138003', '王五', 'dev', 1, 0),
(NOW(), NOW(), '13800138004', '赵六', 'test', 1, 0);

-- 用户配置表
CREATE TABLE IF NOT EXISTS `user_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `dashboard_config` text COMMENT '看板配置(JSON格式)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_id` (`user_id`),
  KEY `idx_user_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 导出任务表
CREATE TABLE IF NOT EXISTS `export_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL COMMENT '创建任务的用户ID',
  `status` varchar(20) DEFAULT 'pending' COMMENT '状态: pending/processing/completed/failed',
  `total_count` int DEFAULT '0' COMMENT '总记录数',
  `process_count` int DEFAULT '0' COMMENT '已处理记录数',
  `file_path` varchar(500) DEFAULT NULL COMMENT '生成的文件路径',
  `filename` varchar(200) DEFAULT NULL COMMENT '文件名',
  `error` text COMMENT '错误信息',
  `completed_at` datetime(3) DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_export_tasks_deleted_at` (`deleted_at`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
