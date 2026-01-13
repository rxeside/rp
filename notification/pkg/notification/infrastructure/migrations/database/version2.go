package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion2(client mysql.ClientContext) migrator.Migration {
	return &version2{
		client: client,
	}
}

type version2 struct {
	client mysql.ClientContext
}

func (v version2) Version() int64 {
	return 2
}

func (v version2) Description() string {
	return "Create 'notification' table"
}

func (v version2) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE notification
		(
		    id          VARCHAR(36)  NOT NULL,
		    payload     JSON         NOT NULL,
		    executed_at DATETIME     NULL,
		    created_at  DATETIME     NOT NULL,
		    updated_at  DATETIME     NOT NULL,
		    status      VARCHAR(32)  NULL,
		    PRIMARY KEY (id)
		)
		    ENGINE = InnoDB
		    CHARACTER SET = utf8mb4
		    COLLATE utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
