package main

import (
	"errors"
	"github.com/gdamore/tcell/v2"
	dockerBuilder "github.com/helmutkemper/iotmaker.docker.builder"
	dockerNetwork "github.com/helmutkemper/iotmaker.docker.builder.network"
	networkInterface "github.com/helmutkemper/iotmaker.docker.builder.network.interface"
	"github.com/rivo/tview"
)

type Terminal struct {
	application *tview.Application
	pages       *tview.Pages
	modal       *tview.Modal

	networkDocker networkInterface.ContainerBuilderNetworkInterface
	builderDocker *dockerBuilder.ContainerBuilder
}

func (e *Terminal) SetNetwork(value networkInterface.ContainerBuilderNetworkInterface) {
	e.networkDocker = value
}

func (e *Terminal) SetBuilder(value *dockerBuilder.ContainerBuilder) {
	e.builderDocker = value
}

func (e *Terminal) Init() (object interface{}, err error) {
	if e.networkDocker == nil {
		err = errors.New("network docker interface is not set")
		return
	}

	if e.networkDocker == nil {
		err = errors.New("network docker interface is not set")
		return
	}

	if e.builderDocker == nil {
		err = errors.New("builder docker interface is not set")
		return
	}

	e.pages = tview.NewPages()
	e.mountNavigationPages()
	e.pages.SendToFront("mainWindow")

	e.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			e.application.Stop()
		}
		return event
	})

	err = e.application.SetRoot(e.pages, true).Run()
	return e, err
}

func (e *Terminal) mountNavigationPages() {

	e.mountPageMain()
	e.pages.AddPage("modal", e.modal, true, true)

	return
}

func (e *Terminal) mountPageMain() {
	var menu = tview.NewList()
	menu.AddItem(
		"Instalar o nats",
		"Baixa e instala o nats",
		'i',
		func() {
			menu.SetCurrentItem(0)
			e.pages.SendToFront("installNats")
		},
	).
		AddItem(
			"Parar o nats",
			"Parar o container nats_delete_after_test",
			'p',
			func() {
				menu.SetCurrentItem(0)
				e.pages.SendToFront("natsStop")
			},
		).
		AddItem(
			"Inicia o nats",
			"Inicia o container nats_delete_after_test",
			'i',
			func() {
				menu.SetCurrentItem(0)
				e.pages.SendToFront("natsStart")
			},
		).
		AddItem(
			"Remover conteúdo",
			"Remove todo o conteúdo de suporte",
			'r',
			func() {
				menu.SetCurrentItem(0)
				e.pages.SendToFront("natsRemoveContainer")
			},
		).
		AddItem(
			"Remover imagens",
			"Remove todas as imagens baixadas",
			'd',
			func() {
				menu.SetCurrentItem(0)
				e.pages.SendToFront("natsRemoveImage")
			},
		).
		AddItem(
			"Monta o project",
			"Monta e instala o projeto",
			'm',
			func() {
				menu.SetCurrentItem(0)
				e.pages.SendToFront("projectBuild")
			},
		).
		AddItem(
			"Quit",
			"Press to exit",
			'q',
			func() {
				menu.SetCurrentItem(0)
				e.application.Stop()
			},
		)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)

	var helper = tview.NewTextView().
		SetText("Este aplicativo dá suporte para instalar e remover todos os containers de suporte necessário para simular o funcionamento da cache.\n\n\nEntre na funcionalidade descrita no menu para instruções detalhadas antes de prosseguir.\n\n\nAperte a tecla `ESC` a qualquer momento para fechar o programa.\n\n\n### Apenas uso interno da TC ###")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)

	e.pages.AddPage("mainWindow", e.mountDefaultViewWithFlex(menu, helper), true, true)
}

func (e *Terminal) mountDefaultViewWithFlex(menu, help tview.Primitive) (flex *tview.Flex) {
	flex = tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(menu, 0, 100, true), 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(help, 0, 100, false), 0, 3, false)

	return
}

func main() {
	var err error

	network := dockerNetwork.ContainerBuilderNetwork{}

	builder := dockerBuilder.ContainerBuilder{}

	t := Terminal{}
	t.SetNetwork(&network)
	t.SetBuilder(&builder)
	_, err = t.Init()
	if err != nil {
		panic(err)
	}
}
