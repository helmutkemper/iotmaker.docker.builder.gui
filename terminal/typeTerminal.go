package terminal

import (
	"errors"
	"github.com/gdamore/tcell/v2"
	dockerBuilder "github.com/helmutkemper/iotmaker.docker.builder"
	dockerNetwork "github.com/helmutkemper/iotmaker.docker.builder.network"
	networkInterface "github.com/helmutkemper/iotmaker.docker.builder.network.interface"
	"github.com/rivo/tview"
	"regexp"
)

type Terminal struct {
	application *tview.Application
	pages       *tview.Pages
	modal       *tview.Modal

	menu    tview.Primitive
	content tview.Primitive

	focus []tview.Primitive

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

	e.application = tview.NewApplication()
	e.application.EnableMouse(true)
	e.application.SetBeforeDrawFunc(
		func(s tcell.Screen) bool {
			s.Clear()
			return false
		},
	)

	e.modal = tview.NewModal()

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
	e.mountGitPageName()
	e.mountGitPage()
	e.pages.AddPage("modal", e.modal, true, true)

	return
}

func (e *Terminal) mountPageMain() {
	var menu = tview.NewList()

	menu.AddItem(
		"Configurações",
		"Mostra as configurações do projeto atual",
		's',
		func() {
			menu.SetCurrentItem(0)
			e.pages.SendToFront("listInfrastructure")
		},
	)

	menu.AddItem(
		"Listar infraestrutura",
		"Lista a infraestrutura atual do teste",
		'a',
		func() {
			menu.SetCurrentItem(1)
			e.pages.SendToFront("listInfrastructure")
		},
	)

	menu.AddItem(
		"Montar imagem",
		"Baixar e montar uma imagem pública contida no docker hub",
		'i',
		func() {
			menu.SetCurrentItem(2)
			e.pages.SendToFront("imageMount")
		},
	)

	menu.AddItem(
		"Projeto em servidor git",
		"Baixar e montar um container baseado no conteúdo de um servidor git",
		'g',
		func() {
			menu.SetCurrentItem(3)
			e.pages.SendToFront("gitProjectQuestion")

		},
	)

	menu.AddItem(
		"Projeto em pasta local",
		"Usa o conteúdo de uma pasta local para montar um container",
		'l',
		func() {
			menu.SetCurrentItem(4)
			e.pages.SendToFront("folderProject")
		},
	)

	menu.AddItem(
		"Gerar código",
		"Gera o código de teste para ser adicionado ao seu projeto",
		'c',
		func() {
			menu.SetCurrentItem(5)
			e.pages.SendToFront("folderProject")
		},
	)

	menu.AddItem(
		"Sair",
		"Pressione para sair",
		'x',
		func() {
			menu.SetCurrentItem(6)
			e.application.Stop()
		},
	)

	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)

	var helper = tview.NewTextView().
		SetText(
			"Este aplicativo dá suporte a criação de teste de caos durante o desenvolvimento de microsserviços." +
				"\n\n\n" +
				"Ele funciona como uma interface gráfica onde você informa quais containers necessita criar, e ao final, um código de teste de caos é gerado para você usar." +
				"\n\n\n" +
				"Comece pela ordem de criação da infraestrutura de teste. Por exemplo, se você necessita do NATS rodando para subir seu projeto, o nats deve ser informado primeiro.",
		)
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)

	e.pages.AddPage("mainWindow", e.mountDefaultViewWithFlex(menu, helper), true, true)
}

func (e *Terminal) mountGitPage() {
	var menu = tview.NewList()

	menu.AddItem(
		"Editar",
		"Editar o conteúdo a direita",
		'e',
		func() {
			menu.SetCurrentItem(0)

		},
	)

	menu.AddItem(
		"Expor porta",
		"Expõe uma porta de rede",
		'e',
		func() {
			menu.SetCurrentItem(1)
			e.pages.SendToFront("listInfrastructure")
		},
	)

	menu.AddItem(
		"Expor e trocar portas",
		"Expõe uma porta de rede trocando o valor",
		't',
		func() {
			menu.SetCurrentItem(0)
			e.pages.SendToFront("listInfrastructure")
		},
	)

	menu.AddItem(
		"Adicionar um volume",
		"Compartilha pasta/arquivo local com o container",
		'v',
		func() {
			menu.SetCurrentItem(2)
			e.pages.SendToFront("imageMount")
		},
	)

	menu.AddItem(
		"Substituir um arquivo",
		"Substitui arquivos antes de montar à imagem",
		'g',
		func() {
			menu.SetCurrentItem(3)
			e.pages.SendToFront("gitProject")
		},
	)

	menu.AddItem(
		"Salvar",
		"Salva a configuração",
		's',
		func() {
			menu.SetCurrentItem(4)
			e.pages.SendToFront("folderProject")
		},
	)

	menu.AddItem(
		"Menu principal",
		"Retorna ao menu principal",
		'x',
		func() {
			menu.SetCurrentItem(5)
			e.pages.SendToFront("mainWindow")
		},
	)

	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)

	var helper = tview.NewForm().
		AddCheckbox("Imprimir a saída padrão do container", true, nil).
		AddCheckbox("Cópia das chaves ssh automática", true, nil).
		AddDropDown("Montar o Dockerfile", []string{"completo", "simples", "não montar"}, 0, nil).
		AddInputField("Repositório privado", "", 20, nil, nil).
		AddInputField("URL do git", "", 20, nil, nil)

	//var helper = tview.NewTextView().
	//		SetText(
	//			"Este aplicativo dá suporte a criação de teste de caos durante o desenvolvimento de microsserviços." +
	//					"\n\n\n" +
	//					"Ele funciona como uma interface gráfica onde você informa quais containers necessita criar, e ao final, um código de teste de caos é gerado para você usar." +
	//					"\n\n\n" +
	//					"Comece pela ordem de criação da infraestrutura de teste. Por exemplo, se você necessita do NATS rodando para subir seu projeto, o nats deve ser informado primeiro.",
	//		)
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)

	e.pages.AddPage("gitProject", e.mountDefaultViewWithFlex(menu, helper), true, true)
}

func (e *Terminal) mountGitPageName() {

	var question = tview.NewInputField().
		SetLabel("Nome: ").
		SetPlaceholder("E.g. nats").
		SetFieldWidth(20).
		SetAcceptanceFunc(e.acceptLowCaseTextOnly).
		SetDoneFunc(
			func(key tcell.Key) {
				if key == tcell.KeyEnter {
					e.pages.SendToFront("gitProject")
				}
			},
		)

	e.pages.AddPage("gitProjectQuestion", e.mountQuestionViewWithFlex(question), true, true)
}

func (e *Terminal) mountDefaultViewWithFlex(menu, help tview.Primitive) (flex *tview.Flex) {
	flex = tview.NewFlex()

	e.menu = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(menu, 0, 100, true)
	e.content = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(help, 0, 100, true)

	flex.AddItem(e.menu, 0, 1, true)
	flex.AddItem(e.content, 0, 3, true)

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			e.cycleFocus(e.application, false, e.focus...)
		} else if event.Key() == tcell.KeyBacktab {
			e.cycleFocus(e.application, true, e.focus...)
		}
		return event
	})

	e.focus = []tview.Primitive{e.content, e.menu}

	return
}

func (e *Terminal) mountQuestionViewWithFlex(question tview.Primitive) (flex *tview.Flex) {

	flex = tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewList(), 0, 100, false), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(question, 0, 100, true), 0, 3, true)

	return
}

func (e *Terminal) acceptLowCaseTextOnly(_ string, ch rune) (accept bool) {
	var re *regexp.Regexp
	re = regexp.MustCompile("[a-z_]")

	return re.Match([]byte{byte(ch)})
}

func (e *Terminal) cycleFocus(app *tview.Application, reverse bool, elements ...tview.Primitive) {
	for i, element := range elements {
		if !element.HasFocus() {
			continue
		}

		length := len(elements) - 1

		if reverse {
			i -= 1
			if i < 0 {
				i = length
			}
		} else {
			i += 1
			if i > length {
				i = 0
			}
		}

		app.SetFocus(elements[i])
		return
	}
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
