package main

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	`.----------------.  .----------------.  .----------------.  .----------------.  .----------------.  .----------------.  .----------------.  .----------------.  .----------------. `,
	`| .--------------. || .--------------. || .--------------. || .--------------. || .--------------. || .--------------. || .--------------. || .--------------. || .--------------. |`,
	`| |  ____  ____  | || |  _________   | || |      __      | || |   _____      | || |  _________   | || |  ____  ____  | || |     ______   | || |  _________   | || |   _____      | |`,
	`| | |_   ||   _| | || | |_   ___  |  | || |     /  \     | || |  |_   _|     | || | |  _   _  |  | || | |_   ||   _| | || |   .' ___  |  | || | |  _   _  |  | || |  |_   _|     | |`,
	`| |   | |__| |   | || |   | |_  \_|  | || |    / /\ \    | || |    | |       | || | |_/ | | \_|  | || |   | |__| |   | || |  / .'   \_|  | || | |_/ | | \_|  | || |    | |       | |`,
	`| |   |  __  |   | || |   |  _|  _   | || |   / ____ \   | || |    | |   _   | || |     | |      | || |   |  __  |   | || |  | |         | || |     | |      | || |    | |   _   | |`,
	`| |  _| |  | |_  | || |  _| |___/ |  | || | _/ /    \ \_ | || |   _| |__/ |  | || |    _| |_     | || |  _| |  | |_  | || |  \ '.___ '\  | || |    _| |_     | || |   _| |__/ |  | |`,
	`| | |____||____| | || | |_________|  | || ||____|  |____|| || |  |________|  | || |   |_____|    | || | |____||____| | || |   '._____.'  | || |   |_____|    | || |  |________|  | |`,
	`| |              | || |              | || |              | || |              | || |              | || |              | || |              | || |              | || |              | |`,
	`| '--------------' || '--------------' || '--------------' || '--------------' || '--------------' || '--------------' || '--------------' || '--------------' || '--------------' |`,
	` '----------------'  '----------------'  '----------------'  '----------------'  '----------------'  '----------------'  '----------------'  '----------------'  '----------------' `,
}

func createApplication() (app *tview.Application) {
	app = tview.NewApplication()
	pages := tview.NewPages()

	infoUI := createInfoPanel(app)
	logPanel := createTextViewPanel(app, "Log")

	log.SetOutput(logPanel)

	clusterList := getClusterList()
	// the clusterList here will be retreived from KubeConfig command.
	// For test purpose, this is hardcoded now.

	clusterList.AddItem("Cluster 1", "", '1', nil)
	clusterList.AddItem("Cluster 2", "", '2', nil)
	clusterList.AddItem("Cluster 3", "", '3', nil)
	clusterList.AddItem("Cluster 4", "", '4', nil)

	commandList := createCommandList()
	commandList.AddItem("K8s Sanity", "", 'k', sendCommand(pages, infoUI))
	commandList.AddItem("Infra Sanity", "", 'i', sendCommand(pages, infoUI))
	commandList.AddItem("PAAS Sanity", "", 'p', sendCommand(pages, infoUI))
	commandList.AddItem("SMF Sanity", "", 'a', sendCommand(pages, infoUI))
	commandList.AddItem("UPF Sanity", "", 'u', sendCommand(pages, infoUI))
	commandList.AddItem("Storage Analysis", "", 's', sendCommand(pages, infoUI))
	commandList.AddItem("Netpol Scan", "", 'n', sendCommand(pages, infoUI))
	commandList.AddItem("Stop", "", 's', stop(infoUI))
	commandList.AddItem("Quit", "", 'q', func() {
		app.Stop()
	})

	reportList := createReportList()
	reportList.AddItem("2024-17-09", "", 0, nil)
	reportList.AddItem("2024-16-09", "", 0, nil)
	reportList.AddItem("2024-15-09", "", 0, nil)

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
	banner.SetTextAlign(tview.AlignLeft)
	banner.SetDynamicColors(true)
	banner.SetTextAlign(tview.AlignCenter)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(clusterList, 0, 30, true).
		AddItem(commandList, 0, 30, true).
		AddItem(reportsPanel, 0, 30, false)

	info := tview.NewTextView()
	info.SetBorder(true)
	info.SetText("HealthCtl v1.0 - Copyright 2024 Microsoft Corp")
	info.SetTextAlign(tview.AlignCenter)

	layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(banner, 13, 1, false).
		AddItem(mainLayout, 20, 30, true).
		AddItem(info, 0, 1, false)

	return layout
}

func sendCommand(pages *tview.Pages, infoUI *testInfoUI) func() {
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
		form.AddButton("Start", startFunc)
		form.AddButton("Cancel", cancelFunc)
		form.SetCancelFunc(cancelFunc)
		form.SetButtonsAlign(tview.AlignCenter)

		form.SetBorder(true).SetTitle(fmt.Sprintf("Send 0x%02X and Listen", "Sample test"))

		modal := createModalForm(pages, form, 13, 55)

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
	app := createApplication()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
