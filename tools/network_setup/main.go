package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/perryd01/vaccination-slot/internal/config"
)

type NetworkFlags struct {
	ContainerName *string
	Reuse         *bool
	HostPort      *uint
	ImageName     *string
	ContainerID   *string
}

func (nf *NetworkFlags) Print() {
	fmt.Println("ContainerName:", *nf.ContainerName, "Reuse:", *nf.Reuse, "HostPort:", *nf.HostPort, "ImageName:", *nf.ImageName, "ContainerID:", *nf.ContainerID)
}

var nf NetworkFlags

func init() {
	var imageName = "ibmcom/ibp-microfab"
	nf.Reuse = flag.Bool("reuse", false, "reuse container if exists")
	nf.HostPort = flag.Uint("hport", 8080, "host port number")
	nf.ContainerName = flag.String("cname", "vacc_slot", "docker ContainerName")
	nf.ImageName = &imageName
}

//go:embed network.tmpl
var networkTmpl string

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
	fmt.Println(buf.String())

	c, err := getPreviousContainer(*nf.ContainerName, *nf.ImageName)
	if err != nil {
		log.Fatal(err)
	}

	if c != nil {
		nf.ContainerID = &c.ID
	}

	nf.Print()

	var cmd *exec.Cmd
	if nf.ContainerID == nil || *nf.Reuse == false {

		err := deleteContainer("", *nf.ContainerName)
		if err != nil {
			log.Fatal(err)
		}

		// docker run -e MICROFAB_CONFIG -p 8080:8080 ibmcom/ibp-microfab
		cmd = exec.Command("docker", "run", "--name", *nf.ContainerName, "-e", "MICROFAB_CONFIG", "-p", strconv.FormatUint(uint64(*nf.HostPort), 10)+":8080", "ibmcom/ibp-microfab")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, buf.String())
	} else {
		fmt.Printf("reusing previous container %s", *nf.ContainerID)
		cmd = exec.Command("docker", "start", *nf.ContainerID, "-i")
		cmd.Env = os.Environ()
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

}

func deleteContainer(ID string, name string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	if ID == "" && name == "" {
		return errors.New("cannot delete container with empty ID and/or Name")
	}

	if ID == "" && name != "" {
		containers, err := getAllContainers()
		if err != nil {
			return err
		}

		// Get ID by name
		for i := 0; i < len(containers); i++ {
			if contains(containers[i].Names, "/"+name) {
				ID = containers[i].ID
				break
			}
		}
	}

	err = cli.ContainerRemove(context.Background(), ID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}

func getAllContainers() ([]types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func getPreviousContainer(containerName string, imageName string) (*types.Container, error) {
	containers, err := getAllContainers()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		// container.Names start with a backslash
		if container.Image == imageName && contains(container.Names, "/"+containerName) {
			return &container, nil
		}
	}
	return nil, nil
}

func contains[T string | int | uint](arr []T, e T) bool {
	for _, element := range arr {
		if element == e {
			return true
		}
	}
	return false
}
