package docker

import (
	"fmt"
	"sync"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	mut      = sync.Mutex{}
	pool     *dockertest.Pool
	resource *dockertest.Resource
)

func RunDocker(imageType Image) string {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("dockertest.NewPool failed: %s", err))
	}

	if err := pool.Client.Ping(); err != nil {
		panic(fmt.Sprintf("pool.Client.Ping failed: %s", err))
	}

	mut.Lock()
	defer mut.Unlock()

	imageInfo, ok := imageInfos[imageType]
	if !ok {
		panic(fmt.Sprintf("imageType %d not found", imageType))
	}

	resource, err = pool.RunWithOptions(&imageInfo.RunOptions,
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		panic(fmt.Sprintf("pool.RunWithOptions failed: %s", err))
	}

	port := resource.GetPort(imageInfo.Port)
	if err := pool.Retry(func() error {
		return imageInfo.CheckReadyFunc(port)
	}); err != nil {
		panic(fmt.Sprintf("pool.Retry failed: %s", err))
	}

	return port
}

func PurgeDocker() {
	if err := pool.Purge(resource); err != nil {
		panic(fmt.Sprintf("pool.Purge failed: %s", err))
	}
}
