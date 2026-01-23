/*
 *@author  chengkenli
 *@project srAnr
 *@package util
 *@file    init
 *@date    2025/6/12 15:47
 */

package util

const SchemaMetaCreate = `
CREATE TABLE SchemaMetaCreate (
  app                varchar(100)  NOT NULL COMMENT 'App StarRocks CN Name',
  nickname           varchar(100)  DEFAULT NULL COMMENT 'Alias',
  alias              varchar(100)  DEFAULT NULL COMMENT 'App Alias',
  feip               varchar(200)  NOT NULL COMMENT 'F5,VIP,CLB,FE',
  user               varchar(200)  NOT NULL COMMENT 'StarRocks Admin User',
  password           varchar(500)  NOT NULL COMMENT 'StarRocks Admin Password',
  feport             int           NOT NULL DEFAULT '9030' COMMENT 'FE Query Port',
  fe_log_path  varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'FE 日志目录',
  be_log_path  varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'BE 日志目录',
  status             int           NOT NULL DEFAULT '0' COMMENT '0 off, 1 on',
  updated_at         timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'CURRENT_TIMESTAMP'
) ENGINE=InnoDB COMMENT='StarRocks信息表';
`

const SchemaMetaInsert = "INSERT INTO SchemaMetaInsert (app, nickname, alias, feip, `user`, password, feport,fe_log_path,be_log_path, status) VALUES('sr-test', 'StarRocks(Tencent Cloud) SR-TEST', NULL, '127.0.0.1', 'chengkenli', '**************1.', 9030,'','', 0)"
