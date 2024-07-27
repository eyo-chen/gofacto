package docker

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ImageMongo: {
		RunOptions: dockertest.RunOptions{
			Repository: "mongo",
			Tag:        "4.4",
		},
		Port: "27017/tcp",
		CheckReadyFunc: func(port string) error {
			ctx := context.Background()
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://localhost:%s", port)))
			if err != nil {
				log.Fatalf("mongo.Connect failed: %s", err)
			}

			return client.Ping(ctx, nil)
		},
	},
}
