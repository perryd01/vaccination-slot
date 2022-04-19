package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/perryd01/vaccination-slot/internal/config"
	"log"
	"os"
	"os/exec"
	"text/template"
)

//go:embed network.tmpl
var networkTmpl string

func main() {
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
	fmt.Println(string(buf.Bytes()))

	// docker run -e MICROFAB_CONFIG -p 8080:8080 ibmcom/ibp-microfab
	cmd := exec.Command("docker", "run", "-e", "MICROFAB_CONFIG", "-p", "8080:8080", "ibmcom/ibp-microfab")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, string(buf.Bytes()))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
