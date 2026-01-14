CREATE TABLE IF NOT EXISTS products
(
    `id`         VARCHAR(36) NOT NULL,
    `name`       VARCHAR(255) NOT NULL,
    `price`      DECIMAL(10,2) NOT NULL,
    `quantity`   INT NOT NULL,
    `created_at` DATETIME     NOT NULL,
    `updated_at` DATETIME     NOT NULL,
    `deleted_at` DATETIME     NULL,
    PRIMARY KEY (`id`)
    ) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci;