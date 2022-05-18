package main

import (
	dockerBuilder "github.com/helmutkemper/iotmaker.docker.builder"
	"github.com/helmutkemper/iotmaker.docker.builder.gui/terminal"
	dockerNetwork "github.com/helmutkemper/iotmaker.docker.builder.network"
)

func main() {
	var err error

	network := dockerNetwork.ContainerBuilderNetwork{}

	builder := dockerBuilder.ContainerBuilder{}

	t := terminal.Terminal{}
	t.SetNetwork(&network)
	t.SetBuilder(&builder)
	_, err = t.Init()
	if err != nil {
		panic(err)
	}
}
