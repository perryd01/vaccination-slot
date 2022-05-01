package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"github.com/docker/go-connections/nat"
	"log"
	"strconv"
	"text/template"

	"github.com/docker/docker/api/types"
	dc "github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/perryd01/vaccination-slot/internal/config"
)

var nf struct {
	ContainerName string
	Reuse         bool
	HostPort      uint
	ImageName     string
	ContainerID   string
}

func init() {
	var imageName = "ibmcom/ibp-microfab"
	flag.BoolVar(&nf.Reuse, "reuse", false, "reuse container if exists")
	flag.UintVar(&nf.HostPort, "hport", 8080, "host port number")
	flag.StringVar(&nf.ContainerName, "cname", "vacc_slot", "docker ContainerName")
	nf.ImageName = imageName
}

//go:embed network.tmpl
var networkTmpl string

var client *docker.Client

func main() {
	flag.Parse()
	tmpl := template.New("network")
	_, err := tmpl.Parse(networkTmpl)
	if err != nil {
		log.Fatal(err)
	}
	n := config.NetworkConfig()
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, &n)
	if err != nil {
		log.Fatal(err)
	}

	client, err = docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	c, err := getPreviousContainer(nf.ContainerName, nf.ImageName)
	if err != nil {
		_ = fmt.Errorf("%v", err)
	} else {
		nf.ContainerID = c.ID
	}

	fmt.Printf("%+v\n", nf)

	if !nf.Reuse || len(nf.ContainerID) == 0 {
		if len(nf.ContainerID) > 0 {
			if err = client.ContainerRemove(context.Background(), nf.ContainerID, types.ContainerRemoveOptions{}); err != nil {
				log.Fatal(err)
			}
		}
		container, err := client.ContainerCreate(context.Background(),
			&dc.Config{
				Image:        nf.ImageName,
				Cmd:          []string{"/tini", "--", "/docker-entrypoint.sh"},
				Env:          []string{buf.String()},
				ExposedPorts: nat.PortSet{"8080": struct{}{}},
			},
			&dc.HostConfig{
				PortBindings: nat.PortMap{"8080": {{HostIP: "0.0.0.0", HostPort: strconv.FormatUint(uint64(nf.HostPort), 10)}}},
			}, nil, nil, nf.ContainerName)
		if err != nil {
			log.Fatal(err)
		}
		nf.ContainerID = container.ID
	}

	if err = client.ContainerStart(context.Background(), nf.ContainerID, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}
}

func getPreviousContainer(containerName string, imageName string) (types.Container, error) {
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return types.Container{}, err
	}

	for _, container := range containers {
		// container.Names start with a backslash
		if container.Image == imageName && contains(container.Names, "/"+containerName) {
			return container, nil
		}
	}
	return types.Container{}, fmt.Errorf("can't find last container")
}

func contains[T string | int | uint](arr []T, e T) bool {
	for _, element := range arr {
		if element == e {
			return true
		}
	}
	return false
}
