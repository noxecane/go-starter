package config

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-pg/pg/v10"
)

func SetupDB(env Env) (*pg.DB, error) {
	opts := &pg.Options{
		Addr:            fmt.Sprintf("%s:%d", env.PostgresHost, env.PostgresPort),
		User:            env.PostgresUser,
		Password:        env.PostgresPassword,
		Database:        env.PostgresDatabase,
		ApplicationName: env.Name,
		OnConnect: func(ctx context.Context, cn *pg.Conn) error {
			if _, err := cn.ExecContext(ctx, "set search_path=?", env.Name); err != nil {
				return err
			}
			return nil
		},
	}

	if env.PostgresSecureMode {
		opts.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	db := pg.Connect(opts)
	_, err := db.Exec("select version()")

	return db, err
}
