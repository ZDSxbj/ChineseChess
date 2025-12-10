-- 添加 ai_level 字段到 game_records 表
-- 用于存储人机对战的 AI 难度（1-6）

ALTER TABLE `game_records` 
ADD COLUMN `ai_level` INT NOT NULL DEFAULT 3 COMMENT 'AI难度: 1-6，仅在 game_type=1 时有效';

-- 为已有的人机对战记录设置默认难度
UPDATE `game_records` 
SET `ai_level` = 3 
WHERE `game_type` = 1 AND `ai_level` = 0;

-- 添加索引优化查询
CREATE INDEX `idx_game_type_ai_level` ON `game_records` (`game_type`, `ai_level`);
