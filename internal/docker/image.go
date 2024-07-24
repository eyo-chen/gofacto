package docker

import (
	"database/sql"
	"fmt"

	"github.com/ory/dockertest/v3"
)

type Image int32

const (
	// ImageUnspecified is a Image of type Unspecified.
	ImageUnspecified Image = iota

	// ImageMySQL is a Image of type MySQL.
	ImageMySQL

	// ImagePostgres is a Image of type Postgres.
	ImagePostgres

	// ImageMongo is a Image of type Mongo.
	ImageMongo
)

type ImageInfo struct {
	dockertest.RunOptions

	Port           string
	CheckReadyFunc func(port string) error
}

var imageInfos = map[Image]ImageInfo{
	ImageMySQL: {
		RunOptions: dockertest.RunOptions{
			Repository: "mysql",
			Tag:        "8.0",
			Env:        []string{"MYSQL_ROOT_PASSWORD=root"},
		},
		Port: "3306/tcp",
		CheckReadyFunc: func(port string) error {
			db, err := sql.Open("mysql", fmt.Sprintf("root:root@(localhost:%s)/mysql?parseTime=true", port))
			if err != nil {
				return err
			}

			return db.Ping()
		},
	},
	ImagePostgres: {
		RunOptions: dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "13",
			Env:        []string{"POSTGRES_PASSWORD=postgres"},
		},
		Port: "5432/tcp",
		CheckReadyFunc: func(port string) error {
			db, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", port))
			if err != nil {
				return err
			}

			return db.Ping()
		},
	},
}
