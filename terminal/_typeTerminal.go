package terminal

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	iotmakerdocker "github.com/helmutkemper/iotmaker.docker/v1.0.1"
	"github.com/rivo/tview"
	"log"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

const KProjectUnderTestPath = "./projectUnderTest/queue"

type Terminal struct {
	natsDocker         interfaces.DockerNats
	projectDocker      interfaces.DockerProject
	networkDocker      interfaces.Network
	application        *tview.Application
	pages              *tview.Pages
	modal              *tview.Modal
	installNatsView    UpdatedElementsInstallNatsView
	installProjectView UpdatedElementsInstallProjectView
	inspectNats        ContainerNatsInspectView
	inspectProject     containerProjectInspectView
}

type UpdatedElementsInstallNatsView struct {
	percentDownload float64
	percentExtract  float64
	statusText      string
	installTextView *tview.TextView
	docker          *chan iotmakerdocker.ContainerPullStatusSendToChannel
	containerReady  *chan bool
}

type UpdatedElementsInstallProjectView struct {
	percentDownload float64
	percentExtract  float64
	statusText      string
	installTextView *tview.TextView
	docker          *chan iotmakerdocker.ContainerPullStatusSendToChannel
	containerReady  *chan bool
}

type ContainerNatsInspectView struct {
	lastData         *tview.TextView
	lastLog          *tview.TextView
	containerInspect *chan bool
}

type containerProjectInspectView struct {
	lastData         *tview.TextView
	lastLog          *tview.TextView
	containerInspect *chan bool
}

func (e *Terminal) SetNatsDocker(natsDocker interfaces.DockerNats) {
	e.natsDocker = natsDocker
}

func (e *Terminal) SetProjectDocker(projectDocker interfaces.DockerProject) {
	e.projectDocker = projectDocker
}

func (e *Terminal) SetNetworkDocker(networkDocker interfaces.Network) {
	e.networkDocker = networkDocker
}

func (e *Terminal) parseNatsInspectToText(inspect iotmakerdocker.ContainerInspect) {
	var ports string
	var portsListToSort = make([]string, 0)
	var lastData string
	
	for port, portBinding := range inspect.Network.Ports {
		if len(portBinding) == 0 {
			continue
		}
		portsListToSort = append(portsListToSort, port.Port()+"/"+port.Proto()+":"+portBinding[0].HostPort+"/"+port.Proto())
	}
	
	sort.Strings(portsListToSort)
	ports = strings.Join(portsListToSort, "\n            ")
	
	lastData += "Container:  " + inspect.State.Status + "\n"
	lastData += "Portas:     " + ports + "\n"
	lastData += "Gateway:    " + inspect.Network.Gateway + "\n"
	lastData += "IPAddress:  " + inspect.Network.IPAddress + "\n"
	lastData += "MacAddress: " + inspect.Network.MacAddress + "\n"
	
	if e.inspectNats.lastData == nil {
		e.application.Stop()
		log.Printf("91 is nil")
	}
	
	e.inspectNats.lastData.SetText(lastData)
	
	if e.inspectNats.lastLog == nil {
		e.application.Stop()
		log.Printf("91 is nil")
	}
	e.inspectNats.lastLog.SetText(e.natsDocker.GetLastLogs())
}

func (e *Terminal) parseProjectInspectToText(inspect iotmakerdocker.ContainerInspect) {
	var ports string
	var portsListToSort = make([]string, 0)
	var lastData string
	
	for port, portBinding := range inspect.Network.Ports {
		if len(portBinding) == 0 {
			continue
		}
		portsListToSort = append(portsListToSort, port.Port()+"/"+port.Proto()+":"+portBinding[0].HostPort+"/"+port.Proto())
	}
	
	sort.Strings(portsListToSort)
	ports = strings.Join(portsListToSort, "\n            ")
	
	lastData += "Container:  " + inspect.State.Status + "\n"
	lastData += "Portas:     " + ports + "\n"
	lastData += "Gateway:    " + inspect.Network.Gateway + "\n"
	lastData += "IPAddress:  " + inspect.Network.IPAddress + "\n"
	lastData += "MacAddress: " + inspect.Network.MacAddress + "\n"
	
	e.inspectProject.lastData.SetText(lastData)
	
	if e.inspectNats.lastLog == nil {
		e.application.Stop()
		log.Printf("91 is nil")
	}
	e.inspectProject.lastLog.SetText(e.projectDocker.GetLastLogs())
}

func (e *Terminal) Init() (object interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
		
		if commonTypes.App != nil {
			commonTypes.App.Stop()
		}
	}()
	
	if e.networkDocker == nil {
		err = errors.New("network docker interface is not set")
		return
	}
	
	if e.natsDocker == nil {
		err = errors.New("nats docker interface is not set")
		return
	}
	
	err = e.natsDocker.Init()
	if err != nil {
		return
	}
	
	if e.projectDocker == nil {
		err = errors.New("project docker interface is not set")
		return
	}
	
	err = e.projectDocker.Init()
	if err != nil {
		return
	}
	
	e.inspectNats.lastData = tview.NewTextView()
	e.inspectNats.lastData.SetScrollable(true)
	e.inspectNats.lastData.SetChangedFunc(func() {
		e.inspectNats.lastData.ScrollToBeginning()
	})
	e.inspectNats.lastLog = tview.NewTextView()
	e.inspectNats.lastLog.SetChangedFunc(func() {
		e.inspectNats.lastLog.ScrollToEnd()
		e.application.Draw()
	})
	
	e.inspectNats.containerInspect = e.natsDocker.GetChannelOnContainerInspect()
	
	e.inspectProject.lastData = tview.NewTextView()
	e.inspectProject.lastData.SetScrollable(true)
	e.inspectProject.lastData.SetChangedFunc(func() {
		e.inspectProject.lastData.ScrollToBeginning()
	})
	e.inspectProject.lastLog = tview.NewTextView()
	e.inspectProject.lastLog.SetChangedFunc(func() {
		e.inspectProject.lastLog.ScrollToEnd()
		e.application.Draw()
	})
	
	e.inspectProject.containerInspect = e.projectDocker.GetChannelOnContainerInspect()
	
	e.installNatsView.containerReady = e.natsDocker.GetChannelOnContainerReady()
	e.installNatsView.docker = e.natsDocker.GetChannelEvent()
	
	e.installProjectView.containerReady = e.projectDocker.GetChannelOnContainerReady()
	e.installProjectView.docker = e.projectDocker.GetChannelEvent()
	
	e.application = tview.NewApplication()
	e.application.EnableMouse(true)
	
	e.modal = tview.NewModal()
	
	e.installNatsView.installTextView = tview.NewTextView().SetText(e.getStatusText("Parada", 0.0, 0.0))
	e.installNatsView.installTextView.SetChangedFunc(func() {
		e.application.Draw()
	})
	
	e.installProjectView.installTextView = tview.NewTextView().SetText(e.getStatusText("Parada", 0.0, 0.0))
	e.installProjectView.installTextView.SetChangedFunc(func() {
		e.application.Draw()
	})
	
	e.pages = tview.NewPages()
	e.mountNavigationPages()
	e.pages.SendToFront("mainWindow")
	
	e.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			e.application.Stop()
		}
		return event
	})
	
	e.backgroundRun()
	
	commonTypes.App = e.application
	
	err = e.application.SetRoot(e.pages, true).Run()
	return e, err
}

func (e *Terminal) getStatusText(estado string, download, extract float64) (text string) {
	text = fmt.Sprintf("Estado da instalação: %v\n", estado)
	text += fmt.Sprintf("Download da imagem:   %2.1f%%\n", download)
	text += fmt.Sprintf("Extração da imagem:   %2.1f%%", extract)
	return
}

func (e *Terminal) backgroundRun() {
	go func() {
		for {
			e.installNatsView.percentDownload = 0.0
			e.installNatsView.percentExtract = 0.0
			
			var text = e.getStatusText(e.installNatsView.statusText, e.installNatsView.percentDownload, e.installNatsView.percentExtract)
			
			if e.installNatsView.installTextView == nil {
				e.application.Stop()
				log.Printf("221 is nil")
			}
			
			e.installNatsView.installTextView.SetText(text)
			
			e.installProjectView.percentDownload = 0.0
			e.installProjectView.percentExtract = 0.0
			
			text = e.getStatusText(e.installProjectView.statusText, e.installProjectView.percentDownload, e.installProjectView.percentExtract)
			e.installProjectView.installTextView.SetText(text)
			
			time.Sleep(500 * time.Millisecond)
		}
	}()
	
	go func() {
		for {
			select {
			
			case <-*e.inspectNats.containerInspect:
				var inspect = e.natsDocker.GetLastInspect()
				e.parseNatsInspectToText(inspect)
			
			case <-*e.installNatsView.containerReady:
				e.updateNatsStatusTextView("Container pronto para uso")
				e.installNatsView.percentDownload = 100.0
				e.installNatsView.percentExtract = 100.0
			
			case status := <-*e.installNatsView.docker:
				
				if status.SuccessfullyBuildImage == true {
					e.updateNatsStatusTextView("Construção da imagem completa")
				}
				
				if status.SuccessfullyBuildContainer == true {
					e.updateNatsStatusTextView("Construção do container completo")
				}
				
				if status.Closed == true {
					e.updateNatsStatusTextView("Iniciando o container")
					e.installNatsView.percentDownload = 100.0
					e.installNatsView.percentExtract = 100.0
				}
				
				if status.Downloading.Percent != 0.0 {
					e.installNatsView.percentDownload = status.Downloading.Percent
				}
				
				if status.Extracting.Percent != 0.0 {
					e.installNatsView.percentExtract = status.Extracting.Percent
				}
			
			case <-*e.inspectProject.containerInspect:
				var inspect = e.projectDocker.GetLastInspect()
				e.parseProjectInspectToText(inspect)
			
			case <-*e.installProjectView.containerReady:
				e.updateProjectStatusTextView("Container pronto para uso")
				e.installProjectView.percentDownload = 100.0
				e.installProjectView.percentExtract = 100.0
			
			case status := <-*e.installProjectView.docker:
				
				if status.SuccessfullyBuildImage == true {
					e.updateProjectStatusTextView("Construção da imagem completa")
				}
				
				if status.SuccessfullyBuildContainer == true {
					e.updateProjectStatusTextView("Construção do container completo")
				}
				
				if status.Closed == true {
					e.updateProjectStatusTextView("Iniciando o container")
					e.installProjectView.percentDownload = 100.0
					e.installProjectView.percentExtract = 100.0
				}
				
				if status.Downloading.Percent != 0.0 {
					e.installProjectView.percentDownload = status.Downloading.Percent
				}
				
				if status.Extracting.Percent != 0.0 {
					e.installProjectView.percentExtract = status.Extracting.Percent
				}
				
			}
		}
	}()
}

func (e *Terminal) updateNatsStatusTextView(value string) {
	e.installNatsView.statusText = value
}

func (e *Terminal) updateProjectStatusTextView(value string) {
	e.installNatsView.statusText = value
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

func (e *Terminal) mountPageNatsInstall() {
	var menu = tview.NewList()
	menu.AddItem(
		"Instalar",
		"Proceguir com a instalação",
		'i',
		func() {
			menu.SetCurrentItem(0)
			e.updateNatsStatusTextView("Instalando")
			
			//não bloqueia a interface gráfica
			go func() {
				var err error
				
				err = e.projectDocker.RemoveContainer()
				if err != nil {
					e.updateNatsStatusTextView("Erro na remoção do projeto")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na remoção do projeto: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
					return
				}
				
				err = e.natsDocker.RemoveAllByNameContains("nats_delete")
				if err != nil {
					e.updateNatsStatusTextView("Erro na remoção do nats")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na remoção do nats: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
					return
				}
				
				err = e.networkDocker.Remove()
				if err != nil {
					e.updateNatsStatusTextView("Erro na remoção da rede")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na instalação: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
				
				err = e.networkDocker.NetworkCreate("cache_delete_after_test", "12.0.0.0/16", "12.0.0.1")
				if err != nil {
					e.updateNatsStatusTextView("Erro na criação da rede")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na criação: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
				
				err = e.natsDocker.Install()
				if err != nil {
					e.updateNatsStatusTextView("Erro na instalação")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na instalação: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
			}()
			
		},
	).
			AddItem(
				"Retornar",
				"Retornar para o menu principal",
				'r',
				func() {
					menu.SetCurrentItem(0)
					e.pages.SendToFront("mainWindow")
				},
			)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	
	var helper = tview.NewTextView().
		SetText("Instala a imagem `nats:latest`, caso a mesma não exista na máquina, e o container `container_nats_delete_after_test`.\n\n\nCaso o container `container_nats_delete_after_test` exista, o mesmo será removido e instalado em seguida.")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)
	
	var installStatus, containerStatus, logs = e.populateNatsInspectView()
	e.pages.AddPage("installNats", e.mountInspectViewWithFlex(menu, installStatus, containerStatus, logs, helper), true, true)
}

func (e *Terminal) mountPageNatsStop() {
	var menu = tview.NewList()
	menu.AddItem(
		"Parar o container",
		"Parar o container e simular uma falha de conexão.",
		'p',
		func() {
			menu.SetCurrentItem(0)
			e.updateNatsStatusTextView("Comando parar enviado.")
			
			//não bloqueia a interface gráfica
			go func() {
				var err error
				err = e.natsDocker.ContainerStop()
				if err != nil {
					e.modal.ClearButtons()
					e.modal.SetText("Erro ao executar: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
			}()
			
		},
	).
			AddItem(
				"Retornar",
				"Retornar para o menu principal",
				'r',
				func() {
					menu.SetCurrentItem(0)
					e.pages.SendToFront("mainWindow")
				},
			)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	
	var helper = tview.NewTextView().
		SetText("Para o container `container_nats_delete_after_test` e simula uma falha de comunicação no mats.")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)
	
	var installStatus, containerStatus, logs = e.populateNatsInspectView()
	e.pages.AddPage("natsStop", e.mountInspectViewWithFlex(menu, installStatus, containerStatus, logs, helper), true, true)
}

func (e *Terminal) mountPageNatsStart() {
	var menu = tview.NewList()
	menu.AddItem(
		"Iniciar o container",
		"Iniciar o container e simular o retorno da conexão.",
		'p',
		func() {
			menu.SetCurrentItem(0)
			e.updateNatsStatusTextView("Comando iniciar enviado.")
			
			//não bloqueia a interface gráfica
			go func() {
				var err error
				err = e.natsDocker.ContainerStart()
				if err != nil {
					e.modal.ClearButtons()
					e.modal.SetText("Erro ao executar: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
			}()
			
		},
	).
			AddItem(
				"Retornar",
				"Retornar para o menu principal",
				'r',
				func() {
					menu.SetCurrentItem(0)
					e.pages.SendToFront("mainWindow")
				},
			)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	
	var helper = tview.NewTextView().
		SetText("Iniciar o container `container_nats_delete_after_test` e simula um retorno da comunicação no mats.")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)
	
	var installStatus, containerStatus, logs = e.populateNatsInspectView()
	e.pages.AddPage("natsStart", e.mountInspectViewWithFlex(menu, installStatus, containerStatus, logs, helper), true, true)
}

func (e *Terminal) mountPageNatsRemoveContainer() {
	var menu = tview.NewList()
	menu.AddItem(
		"Remover o container",
		"Remover o container para desocupar espaço no computador.",
		'p',
		func() {
			menu.SetCurrentItem(0)
			e.updateNatsStatusTextView("Comando de remoção enviado")
			
			//não bloqueia a interface gráfica
			go func() {
				var err error
				err = e.natsDocker.RemoveAllByNameContains("nats_delete")
				if err != nil {
					e.modal.ClearButtons()
					e.modal.SetText("Erro ao executar: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
			}()
			
		},
	).
			AddItem(
				"Retornar",
				"Retornar para o menu principal",
				'r',
				func() {
					menu.SetCurrentItem(0)
					e.pages.SendToFront("mainWindow")
				},
			)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	
	var helper = tview.NewTextView().
		SetText("Iniciar o container `container_nats_delete_after_test` e simula um retorno da comunicação no mats.")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)
	
	var installStatus, containerStatus, logs = e.populateNatsInspectView()
	e.pages.AddPage("natsRemoveContainer", e.mountInspectViewWithFlex(menu, installStatus, containerStatus, logs, helper), true, true)
}

func (e *Terminal) mountPageNatsRemoveImage() {
	var menu = tview.NewList()
	menu.AddItem(
		"Remover o container",
		"Remover o container para desocupar espaço no computador.",
		'p',
		func() {
			menu.SetCurrentItem(0)
			e.updateNatsStatusTextView("Comando de remoção enviado")
			
			//não bloqueia a interface gráfica
			go func() {
				var err error
				err = e.natsDocker.ImageRemove()
				if err != nil {
					e.modal.ClearButtons()
					e.modal.SetText("Erro ao executar: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				} else {
					e.pages.SendToFront("mainWindow")
				}
			}()
			
		},
	).
			AddItem(
				"Retornar",
				"Retornar para o menu principal",
				'r',
				func() {
					menu.SetCurrentItem(0)
					e.pages.SendToFront("mainWindow")
				},
			)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	
	var helper = tview.NewTextView().
		SetText("Iniciar o container `container_nats_delete_after_test` e simula um retorno da comunicação no mats.")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)
	
	var installStatus, containerStatus, logs = e.populateNatsInspectView()
	e.pages.AddPage("natsRemoveImage", e.mountInspectViewWithFlex(menu, installStatus, containerStatus, logs, helper), true, true)
}

func (e *Terminal) mountPageProjectInstall() {
	var menu = tview.NewList()
	menu.AddItem(
		"Instalar",
		"Proceguir com a instalação",
		'i',
		func() {
			menu.SetCurrentItem(0)
			e.updateProjectStatusTextView("Instalando")
			
			//não bloqueia a interface gráfica
			go func() {
				var err error
				err = e.projectDocker.Build(KProjectUnderTestPath, "image_delete_queue_test:latest", "container_delete_queue_1_test")
				if err != nil {
					e.updateProjectStatusTextView("Erro na montagem da imagem")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na instalação: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
					return
				}
				
				err = e.projectDocker.Install()
				if err != nil {
					e.updateProjectStatusTextView("Erro na instalação")
					e.modal.ClearButtons()
					e.modal.SetText("Erro na instalação: " + err.Error()).
						AddButtons([]string{"Retornar", "Fechar o app"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								switch buttonLabel {
								case "Retornar":
									e.pages.SendToFront("mainWindow")
								case "Fechar o app":
									e.application.Stop()
								}
							})
					e.pages.SendToFront("modal")
				}
			}()
			
		},
	).
			AddItem(
				"Retornar",
				"Retornar para o menu principal",
				'r',
				func() {
					menu.SetCurrentItem(0)
					e.pages.SendToFront("mainWindow")
				},
			)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	
	var helper = tview.NewTextView().
		SetText(">> >> >> escrever a ajuda << << <<")
	helper.SetBorder(true)
	helper.SetBorderPadding(1, 1, 1, 1)
	
	var installStatus, containerStatus, logs = e.populateProjectInspectView()
	e.pages.AddPage("projectBuild", e.mountInspectViewWithFlex(menu, installStatus, containerStatus, logs, helper), true, true)
}

func (e *Terminal) mountNavigationPages() {
	
	e.mountPageMain()
	e.mountPageNatsInstall()
	e.mountPageNatsStop()
	e.mountPageNatsStart()
	e.mountPageNatsRemoveContainer()
	e.mountPageNatsRemoveImage()
	e.mountPageProjectInstall()
	
	e.pages.AddPage("modal", e.modal, true, true)
	
	return
}

func (e *Terminal) mountDefaultViewWithFlex(menu, help tview.Primitive) (flex *tview.Flex) {
	flex = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(menu, 0, 100, true), 0, 1, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(help, 0, 100, false), 0, 3, false)
	
	return
}

func (e *Terminal) populateNatsInspectView() (install, container, logs *tview.Flex) {
	if e.installNatsView.installTextView == nil {
		e.application.Stop()
		log.Printf("774 is nil")
	}
	
	if e.inspectNats.lastData == nil {
		e.application.Stop()
		log.Printf("779 is nil")
	}
	
	if e.inspectNats.lastLog == nil {
		e.application.Stop()
		log.Printf("795 is nil")
	}
	
	install = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(e.installNatsView.installTextView, 0, 100, false), 0, 1, false)
	install.SetBorder(true)
	install.SetBorderPadding(1, 1, 1, 1)
	
	container = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(e.inspectNats.lastData, 0, 100, false), 0, 1, false)
	container.SetBorder(true)
	container.SetBorderPadding(1, 1, 1, 1)
	
	logs = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(e.inspectNats.lastLog, 0, 100, false), 0, 1, false)
	logs.SetBorder(true)
	logs.SetBorderPadding(1, 1, 1, 1)
	
	return
}

func (e *Terminal) populateProjectInspectView() (install, container, logs *tview.Flex) {
	if e.installNatsView.installTextView == nil {
		e.application.Stop()
		log.Printf("822 is nil")
	}
	
	if e.inspectNats.lastData == nil {
		e.application.Stop()
		log.Printf("827 is nil")
	}
	
	if e.inspectNats.lastLog == nil {
		e.application.Stop()
		log.Printf("832 is nil")
	}
	
	install = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(e.installProjectView.installTextView, 0, 100, false), 0, 1, false)
	install.SetBorder(true)
	install.SetBorderPadding(1, 1, 1, 1)
	
	container = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(e.inspectProject.lastData, 0, 100, false), 0, 1, false)
	container.SetBorder(true)
	container.SetBorderPadding(1, 1, 1, 1)
	
	logs = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(e.inspectProject.lastLog, 0, 100, false), 0, 1, false)
	logs.SetBorder(true)
	logs.SetBorderPadding(1, 1, 1, 1)
	
	return
}

func (e *Terminal) mountInspectViewWithFlex(menu, install, container, logs, help tview.Primitive) (flex *tview.Flex) {
	flex = tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(menu, 0, 3, true), 0, 1, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(help, 0, 2, false).
				AddItem(install, 7, 0, false).
				AddItem(container, 12, 0, false).
				AddItem(logs, 0, 6, false), 0, 3, false)
	
	return
}
