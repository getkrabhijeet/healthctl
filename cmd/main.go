package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"healthctl/pkg/k8s"
	"healthctl/pkg/models"
	"healthctl/pkg/testsuite"

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
	logPanel := createTextViewPanel(app, "Output Terminal")
	logPanel.SetDynamicColors(true)

	//do not print date and time
	log.SetFlags(0)
	log.SetOutput(logPanel)
	log.Println("[Green:bl]Welcome to HealthCtl[-:-:-:-]")
	//Print release version and usage information
	log.Println("[Green:b]HealthCtl v1.0[-:-:-:-]")
	//print that healthctl is a tool to run sanity checks on k8s clusters
	log.Println("[Green:b]HealthCtl is a tool to run sanity checks on k8s clusters[-:-:-:-]")
	//print that healthctl is a tool to run sanity checks on application NFs in k8s clusters
	log.Println("[Green:b]HealthCtl is a tool to run sanity checks on application NFs in k8s clusters[-:-:-:-]")
	//check alerts
	log.Println("[Green:b]Check Alerts.[-:-:-:-]")
	//check reports
	log.Println("[Green:b]Check Reports.[-:-:-:-]")
	//check SMF status
	log.Println("[Green:b]Check SMF Status.[-:-:-:-]")
	//check UPF status
	log.Println("[Green:b]Check UPF Status.[-:-:-:-]")
	//check Redis status
	log.Println("[Green:b]Check Redis Status.[-:-:-:-]")
	//collect Kargo
	log.Println("[Green:b]Collecting Kargo[-:-:-:-]")
	//set debug level
	log.Println("[Green:b]Setting Debug Level.[-:-:-:-]")
	//flush Redis
	log.Println("[Green:b]Flushing Redis.[-:-:-:-]")

	kc, _ := k8s.NewK8sClient()
	kc.GetClusterInfo()
	clusterList := createList("Clusters")
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

	commandList := createList("Operations")
	commandList.AddItem("K8s Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Infra Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("PAAS Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("SMF Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("UPF Sanity", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Storage Analysis", "", 0, sendCommand(pages, infoUI, clusterList, commandList))
	commandList.AddItem("Netpol Scan", "", 0, sendCommand(pages, infoUI, clusterList, commandList))

	// create an observation flex with buttons to check alerts
	afn_tools := tview.NewFlex().SetDirection(tview.FlexRow)
	afn_tools.AddItem(tview.NewButton("Alerts").SetSelectedFunc(Alerts(pages)), 3, 1, false)
	afn_tools.AddItem(tview.NewButton("SMF Status").SetSelectedFunc(func() {}), 3, 1, false)
	afn_tools.AddItem(tview.NewButton("UPF Status").SetSelectedFunc(func() {}), 3, 1, false)
	afn_tools.AddItem(tview.NewButton("Redis status").SetSelectedFunc(RedisStatus(pages)), 3, 1, false)
	afn_tools.AddItem(tview.NewButton("Collect Kargo").SetSelectedFunc(func() {}), 3, 1, false)
	afn_tools.AddItem(tview.NewButton("Set Debug Level").SetSelectedFunc(func() {}), 3, 1, false)
	afn_tools.AddItem(tview.NewButton("Flush Redis").SetSelectedFunc(FlushRedis(pages)), 3, 1, false)

	// Set focus navigation
	clusterList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(commandList)
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(logPanel)
			return nil
		}
		return event
	})

	commandList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(logPanel)
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(clusterList)
			return nil
		}
		return event
	})

	logPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(clusterList)
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(commandList)
			return nil
		}
		return event
	})

	layout := createMainLayout(infoUI, clusterList, commandList, logPanel, afn_tools)
	pages.AddPage("main", layout, true, true)

	app.SetRoot(pages, true).EnableMouse(true)

	return app
}

func RedisStatus(pages *tview.Pages) func() {
	kc, _ := k8s.NewK8sClient()
	return func() {
		clearLogPanel(pages)
		redisStatus := kc.GetRedisStatus()
		displayRedisStatus(redisStatus)

	}
}

func FlushRedis(pages *tview.Pages) func() {
	kc, _ := k8s.NewK8sClient()
	return func() {
		clearLogPanel(pages)
		size := kc.GetRedisDbSize()
		for _, s := range size {
			log.Printf("%s : %s \n", s.PodName, s.Output)
		}
		log.Printf("[red:bl]Flushing Redis Data[-:-:-:-]\n")
		err := kc.FlushRedisData()
		if err != nil {
			log.Printf("[red:bl]Error Flushing Redis Data: %v[-:-:-:-]\n", err)
		}
		size = kc.GetRedisDbSize()
		for _, s := range size {
			log.Printf("%s : %s \n", s.PodName, s.Output)
		}
	}
}

func Alerts(pages *tview.Pages) func() {
	kc, _ := k8s.NewK8sClient()
	return func() {
		clearLogPanel(pages)
		alertList := kc.GetAlerts()
		if alertList == nil {
			log.Println("[red]Unable to get alerts[-]")
			return
		}
		displayAlerts(alertList)
	}
}

func displayRedisStatus(r k8s.RedisStatus) {
	var hyphenFormatter = func() {
		log.Printf("| %-33s | %-15s | %-40s | %-50s | %-10s | %-10s | %-15s |\n", "─────────────────────────────────", "───────────────", "────────────────────────────────────────", "──────────────────────────────────────────────────", "──────────", "──────────", "───────────────")
	}
	var newHyphenFormatter = func() {
		log.Printf("| %-155s |\n", "───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	}

	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("Number of Primaries Configured: %d", r.PrimariesConfigured))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("Number of Replicas Configured: %d", r.ReplicasConfigured))
	newHyphenFormatter()

	if r.PodStatus {
		log.Printf("| %-191s |\n", "---- All redis cluster pods are in n/n ready state ----")
	} else {
		log.Printf("| %-191s |\n", "---- Redis cluster pods are [red]NOT[-] in n/n ready state ----")
	}
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("cluster_state: %s", r.ClusterState))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("cluster_slots_ok: %d", r.ClusterSlotsOk))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("cluster_known_nodes: %d", r.ClusterKnownNodes))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("cluster_size: %d", r.ClusterSize))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("cluster_slots_pfail: %d", r.ClusterSlotsPfail))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("cluster_slots_fail: %d", r.ClusterSlotsFail))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("Number of Active zones : %d", r.NumberActiveZones))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("Number of zones where Redis primaries are present : %d", r.NumberZonesPrimaries))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", fmt.Sprintf("Number of Redis primaries on Zone %s : %d", "TODO", r.NumberPrimariesInZone))
	newHyphenFormatter()
	log.Printf("| %-191s |\n", "---- Redis Cluster is working as expected ----")
	hyphenFormatter()
	log.Printf("| %-33s | %-15s | %-40s | %-50s | %-10s | %-10s | %-15s |\n", "PodName", "PodIp", "RedisNodeId", "WorkerNode", "Zone", "CPU", "Memory")
	hyphenFormatter()
	//print all node fields in a table
	for _, node := range r.RedisNodeDetails {
		log.Printf("| %-33s | %-15s | %-40s | %-50s | %-10s | %-10s | %-15s |\n", node.PodName, node.IP, node.ID, r.PodDetails[node.PodName].Worker, node.Zone, r.PodDetails[node.PodName].CPU, r.PodDetails[node.PodName].Memory)
	}
	hyphenFormatter()

}

func displayAlerts(alertList []k8s.Alert) {
	equalFormatter := func() {
		log.Printf("| %-33s | %-8s | %-24s | %-35s | %-30s\n", "=================================", "========", "========================", "===================================", "==============================")
	}
	hyphenFormatter := func() {
		log.Printf("| %-33s | %-8s | %-24s | %-35s | %-30s\n", "---------------------------------", "--------", "------------------------", "-----------------------------------", "------------------------------")
	}

	//clear logPanel

	equalFormatter()
	log.Printf("| %-33s | %-8s | %-24s | %-35s | %-30s\n", "Alertname", "Severity", "Starts At", "Pod Name", "Summary")
	equalFormatter()

	for _, alert := range alertList {
		// if alert.Severity == "critical" {
		// 	alert.Severity = "[red]" + alert.Severity + "[-]"
		// } else if alert.Severity == "major" {
		// 	alert.Severity = "[yellow]" + alert.Severity + "[-]"
		// } else {
		// 	alert.Severity = "[green]" + alert.Severity + "[-]"
		// }

		log.Printf("| %-33s | %-8s | %-24s | %-35s | %-30s\n", alert.AlertName, alert.Severity, alert.StartsAt, alert.PodName, alert.Summary)
		hyphenFormatter()
	}
	log.Printf("| %-33s | %-8s | %-24s | %-35s | %-30s\n", "", "", "", "Total Alerts", strconv.Itoa(len(alertList)))
	equalFormatter()
}

func createMainLayout(infoUI *testInfoUI, clusterList, commandList tview.Primitive, output tview.Primitive, afn_tools *tview.Flex) (layout *tview.Flex) {
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

	commands.SetCellSimple(3, 0, "View Alerts : ")
	commands.GetCell(3, 0).SetAlign(tview.AlignLeft)
	//infoUI.Pods = tview.NewTableCell("none")
	commands.SetCell(3, 1, tview.NewTableCell("a"))

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
		banner.SetCell(i+1, 0, tview.NewTableCell(Logo[i]))
		banner.GetCell(i+1, 0).SetAlign(tview.AlignRight).SetBackgroundColor(tcell.ColorGreen)
	}

	mainLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(clusterList, 20, 0, true).
		AddItem(commandList, 20, 0, true).
		AddItem(output, 0, 30, false).
		AddItem(afn_tools, 20, 0, false)

	info := tview.NewTextView()
	info.SetBorder(true)
	info.SetText("HealthCtl v1.0 - © Microsoft 2024")
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

func clearLogPanel(pages *tview.Pages) {
	_, layout := pages.GetFrontPage()
	layout.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(0).(*tview.Flex).GetItem(2).(*tview.TextView).Clear()
}

func runTests(selectedCluster string, selectedCommand string) {
	kc, _ := k8s.NewK8sClient()
	config := k8s.GetClustersFromKubeConfig()
	kc.SetContext(config, selectedCluster)
	rl := []models.ResourceCheck{}
	if selectedCommand == "K8s Sanity" {
		log.Println("Running K8s Sanity")
		rl = testsuite.CheckK8s(kc.Client)
	} else if selectedCommand == "Infra Sanity" {
		log.Println("Running INFRA Sanity")
		rl = testsuite.CheckINFRA(kc.Client)
	} else if selectedCommand == "PAAS Sanity" {
		log.Println("Running PAAS Sanity")
		rl = testsuite.CheckPAAS(kc.Client)
	} else if selectedCommand == "SMF Sanity" {
		log.Println("Running SMF Sanity")
		rl = testsuite.CheckSMF(kc.Client)
	} else if selectedCommand == "UPF Sanity" {
		log.Println("Running UPF Sanity")
	} else {
		log.Printf("Please select a test to run")
	}

	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
	log.Printf("| %-5s | %-150s | %-7s |\n", "No.", "Test Summary", "Result")
	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")

	status := ""
	for index, resc := range rl {
		if resc.Status {
			status = "PASS"
		} else {
			status = "FAIL"
		}
		log.Printf("| %-5s | %-150s | %-7s |\n", strconv.Itoa(index+1), resc.Details, status)
		log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
	}
	log.Printf("| %-5s | %-150s | %-7s |\n", "", "Total Tests", strconv.Itoa(len(rl)))
	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
	for index, resc := range rl {

		if resc.Status {
			status = "PASS"
		} else {
			status = "FAIL"
		}
		log.Printf("| %-5s | %-150s | %-7s |\n", strconv.Itoa(index+1), resc.Details, status)
		log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
	}
	log.Printf("| %-5s | %-150s | %-7s |\n", "", "Total Tests", strconv.Itoa(len(rl)))
	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
}

func sendCommand(pages *tview.Pages, infoUI *testInfoUI, clusterList *tview.List, commandList *tview.List) func() {
	return func() {
		selectedClusterIndex := clusterList.GetCurrentItem()
		selectedCluster, _ := clusterList.GetItemText(selectedClusterIndex)
		selectedCommandIndex := commandList.GetCurrentItem()
		selectedCommand, _ := commandList.GetItemText(selectedCommandIndex)

		startFunc := func(selectedCluster string, selectedCommand string) {
			stop(infoUI)()
			pages.SwitchToPage("main")
			clearLogPanel(pages)
			runTests(selectedCluster, selectedCommand)
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
			startFunc(selectedCluster, selectedCommand)
		})
		form.AddButton("Cancel", cancelFunc)
		form.SetCancelFunc(cancelFunc)
		form.SetButtonsAlign(tview.AlignCenter)

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
	panel.SetDynamicColors(true)
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

func createList(title string) (newList *tview.List) {
	newList = tview.NewList()
	newList.SetBorder(true).SetTitle(title)
	newList.ShowSecondaryText(false)
	return newList
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
