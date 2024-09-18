package main

import (
	"context"
	"fmt"
	"log"

	"healthctl/pkg/k8s"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type testInfoUI struct {
	ctx       context.Context
	cancel    context.CancelFunc
	app       *tview.Application
	panel     *tview.Flex
	testType  *tview.TextView
	txAddress *tview.TableCell
	cluster   *tview.TableCell
	context   *tview.TableCell
	nodes     *tview.TableCell
	apiserver *tview.TableCell
}

var Logo = []string{
	`┓     ┓ ┓   ┓`,
	`┣┓┏┓┏┓┃╋┣┓┏╋┃`,
	`┛┗┗ ┗┻┗┗┛┗┗┗┗`,
}

var Logo1 = []string{

	` _                _ _   _          _   _ `,
	`| |              | | | | |        | | | |`,
	`| |__   ___  __ _| | |_| |__   ___| |_| |`,
	`| '_ \ / _ \/ _' | | __| '_ \ / __| __| |`,
	`| | | |  __/ (_| | | |_| | | | (__| |_| |`,
	`|_| |_|\___|\__,_|_|\__|_| |_|\___|\__|_|`,
	`                                         `,
}

func createApplication() (app *tview.Application) {
	app = tview.NewApplication()
	pages := tview.NewPages()
	infoUI := createInfoPanel(app)
	logPanel := createTextViewPanel(app, "Log")

	log.SetOutput(logPanel)

	kc, _ := k8s.NewK8sClient()
	kc.GetClusterInfo()
	clusterList := getClusterList()
	config := k8s.GetClustersFromKubeConfig()
	for index, _ := range config.Clusters {
		clusterList.AddItem(index, "", 0, nil)
	}
	handler := func(index int, mainText, secondaryText string, shortcut rune) {
		kc.SetContext(config, mainText)
		infoUI.context.SetText(config.CurrentContext)
		infoUI.cluster.SetText(mainText)
		nodes := kc.GetClusterNodes()
		infoUI.nodes.SetText(fmt.Sprintf("Master: %d, Worker: %d", nodes[0], nodes[1]))
		infoUI.apiserver.SetText(config.Clusters[mainText].Server)
		pages.SwitchToPage("main")
	}
	clusterList.SetChangedFunc(handler)

	commandList := createCommandList()
	commandList.AddItem("K8s Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Infra Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("PAAS Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("SMF Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("UPF Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Storage Analysis", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Netpol Scan", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Stop", "", 's', stop(infoUI))
	commandList.AddItem("Quit", "", 'q', func() {
		app.Stop()
	})

	reportList := createReportList()
	// reportList.AddItem("2024-17-09", "", 0, nil)
	// reportList.AddItem("2024-16-09", "", 0, nil)
	// reportList.AddItem("2024-15-09", "", 0, nil)

	// Set focus navigation
	clusterList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(commandList)
			return nil
		}
		return event
	})

	commandList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(clusterList)
			return nil
			// case tcell.KeyBacktab:
			// 	app.SetFocus(clusterList)
			// 	return nil
		}
		return event
	})

	reportList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBacktab:
			app.SetFocus(commandList)
			return nil
		}
		return event
	})

	layout := createMainLayout(infoUI, clusterList, commandList, reportList)
	pages.AddPage("main", layout, true, true)

	app.SetRoot(pages, true)

	return app
}

func createMainLayout(infoUI *testInfoUI, clusterList, commandList tview.Primitive, reportsPanel tview.Primitive) (layout *tview.Flex) {
	///// Main Layout /////
	metadata := tview.NewTable()
	metadata.SetBorder(true).SetTitle("Cluster Details")

	metadata.SetCellSimple(0, 0, "Context : ")
	metadata.GetCell(0, 0).SetAlign(tview.AlignLeft)
	infoUI.context = tview.NewTableCell("none")
	metadata.SetCell(0, 1, infoUI.context)

	metadata.SetCellSimple(1, 0, "Cluster : ")
	metadata.GetCell(1, 0).SetAlign(tview.AlignLeft)
	infoUI.cluster = tview.NewTableCell("none")
	metadata.SetCell(1, 1, infoUI.cluster)

	metadata.SetCellSimple(2, 0, "Nodes : ")
	metadata.GetCell(2, 0).SetAlign(tview.AlignLeft)
	infoUI.nodes = tview.NewTableCell("Master : 0, Worker : 0")
	metadata.SetCell(2, 1, infoUI.nodes)

	metadata.SetCellSimple(3, 0, "apiserver : ")
	metadata.GetCell(3, 0).SetAlign(tview.AlignLeft)
	infoUI.apiserver = tview.NewTableCell("0")
	metadata.SetCell(3, 1, infoUI.apiserver)

	commands := tview.NewTable()
	commands.SetBorder(true).SetTitle("Shortcuts")

	commands.SetCellSimple(0, 0, "Run Tests : ")
	commands.GetCell(0, 0).SetAlign(tview.AlignLeft)
	//infoUI.Context = tview.NewTableCell("none")
	commands.SetCell(0, 1, tview.NewTableCell("ctrl+r"))

	commands.SetCellSimple(1, 0, "Stop Tests :  ")
	commands.GetCell(1, 0).SetAlign(tview.AlignLeft)
	//infoUI.Cluster = tview.NewTableCell("none")
	commands.SetCell(1, 1, tview.NewTableCell("ctrl+s"))

	commands.SetCellSimple(2, 0, "Open Reports : ")
	commands.GetCell(2, 0).SetAlign(tview.AlignLeft)
	//infoUI.Nodes = tview.NewTableCell("none")
	commands.SetCell(2, 1, tview.NewTableCell("ctrl+o"))

	commands.SetCellSimple(3, 0, "View Topology : ")
	commands.GetCell(3, 0).SetAlign(tview.AlignLeft)
	//infoUI.Pods = tview.NewTableCell("none")
	commands.SetCell(3, 1, tview.NewTableCell("ctrl+t"))

	commands.SetCellSimple(4, 0, "Popeye : ")
	commands.GetCell(4, 0).SetAlign(tview.AlignLeft)
	//infoUI.Pods = tview.NewTableCell("none")
	commands.SetCell(4, 1, tview.NewTableCell("ctrl+p"))

	commands.SetCellSimple(5, 0, "SecurityContexts : ")
	commands.GetCell(5, 0).SetAlign(tview.AlignLeft)
	//infoUI.Pods = tview.NewTableCell("none")
	commands.SetCell(5, 1, tview.NewTableCell("none"))

	banner := tview.NewTable()
	banner.SetBorder(false)
	for i := 0; i < 7; i++ {
		banner.SetCell(i+1, 0, tview.NewTableCell(Logo1[i]))
		banner.GetCell(i+1, 0).SetAlign(tview.AlignRight).SetBackgroundColor(tcell.ColorGreen)
	}

	mainLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(clusterList, 30, 0, true).
		AddItem(commandList, 30, 0, true).
		AddItem(reportsPanel, 0, 30, false)

	info := tview.NewTextView()
	info.SetBorder(true)
	info.SetText("HealthCtl v1.0 - Copyright 2024 Microsoft Corp")
	info.SetTextAlign(tview.AlignCenter)

	header := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(metadata, 0, 1, false).
		AddItem(commands, 0, 1, false).
		AddItem(banner, 43, 1, false)

	mainMenu := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainLayout, 0, 80, true)

	footer := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 0, 1, false)

	layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 9, 1, false).
		AddItem(mainMenu, 0, 1, true).
		AddItem(footer, 3, 1, false)

	return layout
}

func sendCommand(pages *tview.Pages, infoUI *testInfoUI, clusterList *tview.List, commandList *tview.List) func() {
	return func() {
		startFunc := func() {
			stop(infoUI)()
			pages.SwitchToPage("main")
			pages.RemovePage("modal")
			ctx, cancel := context.WithCancel(context.Background())
			infoUI.ctx = ctx
			infoUI.cancel = cancel
			go func() {
				defer func() {
					cancel()
					infoUI.ctx = nil
				}()
				fmt.Println("In sendCommand")
			}()
		}

		cancelFunc := func() {
			pages.SwitchToPage("main")
			pages.RemovePage("modal")
		}

		form := tview.NewForm()
		form.AddButton("Start", func() {
			startFunc()
		})
		form.AddButton("Cancel", cancelFunc)
		form.SetCancelFunc(cancelFunc)
		form.SetButtonsAlign(tview.AlignCenter)

		selectedClusterIndex := clusterList.GetCurrentItem()
		selectedCluster, _ := clusterList.GetItemText(selectedClusterIndex)
		selectedCommandIndex := commandList.GetCurrentItem()
		selectedCommand, _ := commandList.GetItemText(selectedCommandIndex)

		form.SetBorder(true).SetTitle("Confirmation")
		form.AddTextView(fmt.Sprintf("Executing %s command on %s cluster", selectedCommand, selectedCluster), "", 0, 1, false, false)

		modal := createModalForm(pages, form, 13, 80)

		pages.AddPage("modal", modal, true, true)

	}
}

func createInfoPanel(app *tview.Application) (infoUI *testInfoUI) {
	///// Info /////
	infoPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	infoUI = &testInfoUI{}
	infoUI.app = app
	infoUI.panel = infoPanel

	infoUI.testType = tview.NewTextView()
	infoUI.testType.SetBorder(true)
	infoUI.testType.SetText("healthctl")
	infoUI.testType.SetTextAlign(tview.AlignCenter)
	infoPanel.AddItem(infoUI.testType, 0, 1, false)

	txInfo := tview.NewTable()
	txInfo.SetBorder(true).SetTitle("Clusters")

	txInfo.SetCellSimple(0, 0, "Clusters:")
	txInfo.GetCell(0, 0).SetAlign(tview.AlignRight)
	txInfo.SetCell(0, 1, infoUI.txAddress)

	infoInnerPanel := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(txInfo, 0, 1, false)
	infoPanel.AddItem(infoInnerPanel, 0, 1, false)

	return infoUI
}

func createTextViewPanel(app *tview.Application, name string) (panel *tview.TextView) {
	panel = tview.NewTextView()
	panel.SetBorder(true).SetTitle(name)
	panel.SetChangedFunc(func() {
		app.Draw()
	})
	return panel
}

func stop(infoUI *testInfoUI) func() {
	return func() {
		if infoUI.cancel != nil {
			infoUI.cancel()
		}
	}
}

func createCommandList() (commandList *tview.List) {
	///// Commands /////
	commandList = tview.NewList()
	commandList.SetBorder(true).SetTitle("Operation")
	commandList.ShowSecondaryText(false)
	return commandList
}
func getClusterList() (clusterList *tview.List) {
	///// Clusters /////
	clusterList = tview.NewList()
	clusterList.SetBorder(true).SetTitle("Clusters")
	clusterList.ShowSecondaryText(false)
	return clusterList
}

func createReportList() (reportList *tview.List) {
	///// Reports /////
	reportList = tview.NewList()
	reportList.SetBorder(true).SetTitle("Test Results")
	reportList.ShowSecondaryText(false)
	return reportList
}

func createModalForm(pages *tview.Pages, form tview.Primitive, height int, width int) tview.Primitive {
	modal := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
	return modal
}

func main() {
	// kc, _ := k8s.NewK8sClient()
	// r := []paas.ResourceCheck{
	// 	paas.CheckGrafana(kc.Client),
	// 	paas.CheckPrometheus(kc.Client),
	// 	paas.CheckKibana(kc.Client),
	// 	paas.CheckElastic(kc.Client),
	// 	paas.CheckJaeger(kc.Client),
	// 	paas.CheckKiali(kc.Client),
	// 	paas.CheckIstio(kc.Client),
	// 	paas.CheckDbEtcd(kc.Client),
	// 	paas.CheckKubeProm(kc.Client),
	// 	paas.CheckRedisOperator(kc.Client),
	// 	paas.CheckRedisCluster(kc.Client),
	// 	paas.CheckElastAlert(kc.Client),
	// 	paas.CheckAlerta(kc.Client),
	// }

	// paas.PrintResults(r)

	app := createApplication()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
