CREATE DATABASE IF NOT EXISTS demo DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;

use demo;

CREATE TABLE `demo`
(
    `id`          int          NOT NULL AUTO_INCREMENT,
    `name`        VARCHAR(24)  NOT NULL COMMENT '记录名字',
    `desc`        VARCHAR(128) NOT NULL COMMENT '描述信息',
    `create_time` datetime              DEFAULT CURRENT_TIMESTAMP,
    `update_time` datetime              DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `del_flag`    TINYINT      NOT NULL DEFAULT 1 COMMENT '记录是否有效:0=有效;1=删除',
    PRIMARY KEY (`id`),
    KEY `uniq_idx_name` (`name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='demo';


INSERT INTO demo.demo (`name`, `desc`, `del_flag`)
VALUES ('demo', 'this is a demo desc', 0);

CREATE TABLE `user1`
(
    `id`          int          NOT NULL AUTO_INCREMENT,
    `user_id`     INT UNSIGNED NOT NULL COMMENT '用户id',
    `user_name`   VARCHAR(24)  NOT NULL COMMENT '昵称',
    `user_phone`  VARCHAR(30)  NOT NULL DEFAULT '' COMMENT '用户手机号',
    `age`         TINYINT      NOT NULL COMMENT '年龄',
    `create_time` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_uid` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户表';

CREATE TABLE `user2`
(
    `id`          int          NOT NULL AUTO_INCREMENT,
    `user_id`     INT UNSIGNED NOT NULL COMMENT '用户id',
    `user_name`   VARCHAR(24)  NOT NULL COMMENT '昵称',
    `user_phone`  VARCHAR(30)  NOT NULL DEFAULT '' COMMENT '用户手机号',
    `age`         TINYINT      NOT NULL COMMENT '年龄',
    `create_time` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_uid` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户表';

INSERT INTO demo.user1 (`user_id`, `user_name`, `age`)
VALUES (1, 'lily', 10),
       (3, 'jack', 20),
       (5, 'tom', 30);
INSERT INTO demo.user2 (`user_id`, `user_name`, `age`)
VALUES (2, 'lily', 10),
       (4, 'jack', 20),
       (6, 'tom', 30);
