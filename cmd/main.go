package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"healthctl/pkg/k8s"
	"github.com/rivo/tview"
)

type testInfoUI struct {
	ctx       context.Context
	cancel    context.CancelFunc
	app       *tview.Application
	panel     *tview.Flex
	testType  *tview.TextView
	txAddress *tview.TableCell
}

var Logo = []string{
	`┓     ┓ ┓   ┓`,
	`┣┓┏┓┏┓┃╋┣┓┏╋┃`,
	`┛┗┗ ┗┻┗┗┛┗┗┗┗`,
}

func createApplication() (app *tview.Application) {
	app = tview.NewApplication()
	pages := tview.NewPages()

	//kc := k8s.NewK8sClient()
	clusterList := getClusterList()
	for index, _ := range k8s.GetClustersFromKubeConfig() {
		clusterList.AddItem(fmt.Sprintf("%s", index), "", 'x', nil)
	}

	infoUI := createInfoPanel(app)
	logPanel := createTextViewPanel(app, "Log")

	log.SetOutput(logPanel)


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
	reportList.AddItem("2024-17-09", "", 0, nil)
	reportList.AddItem("2024-16-09", "", 0, nil)
	reportList.AddItem("2024-15-09", "", 0, nil)

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
			app.SetFocus(reportList)
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(clusterList)
			return nil
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

	layout := createMainLayout(clusterList, commandList, reportList)
	pages.AddPage("main", layout, true, true)

	app.SetRoot(pages, true)

	return app
}

func createMainLayout(clusterList, commandList tview.Primitive, reportsPanel tview.Primitive) (layout *tview.Flex) {
	///// Main Layout /////
	banner := tview.NewTextView()
	banner.SetBorder(true)
	banner.SetText(strings.Join(Logo, fmt.Sprintf("\n[%s::b]", "green")))
	banner.SetTextAlign(tview.AlignRight)
	banner.SetDynamicColors(true)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(clusterList, 0, 30, true).
		AddItem(commandList, 0, 30, true).
		AddItem(reportsPanel, 0, 30, false)

	info := tview.NewTextView()
	info.SetBorder(true)
	info.SetText("HealthCtl v1.0 - Copyright 2024 Microsoft Corp")
	info.SetTextAlign(tview.AlignCenter)

	layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(banner, 0, 15, false).
		AddItem(mainLayout, 0, 80, true).
		AddItem(info, 0, 5, false)

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
	reportList.SetBorder(true).SetTitle("Reports")
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

	kc, error := k8s.NewK8sClient()
	if error != nil {
		panic(error)
	}
	t := k8s.TestStatus{}

	t = kc.CheckNodes()
	fmt.Printf("Status of Worker nodes : %t\n", t.Status)
	if t.Status == false {
		fmt.Printf("DEBUG: %s, Error: %s", t.Info, t.Error)
	}

	t = kc.CheckPods()
	fmt.Printf("Status of Pods : %t\n", t.Status)
	if t.Status == false {
		fmt.Printf("DEBUG: %s\n", t.Info)
	}

	// for cluster := range k8s.GetClustersFromKubeConfig() {
	// 	fmt.Printf("Cluster: %s\n", cluster)
	// }

	t = kc.CheckEvents()
	fmt.Printf("Status of Events : %t\n", t.Status)
	if t.Status == false {
		fmt.Printf("DEBUG: %s\n", t.Info)
	}

	app := createApplication()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
