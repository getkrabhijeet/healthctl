package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

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

var HEALTH_K8s = "K8s health"
var HEALTH_INFRA = "Infra health"
var HEALTH_PAAS = "PAAS health"
var HEALTH_SMF = "SMF health"
var HEALTH_UPF = "UPF health"
var HEALTH_STORAGE = "Storage health"
var ACTIVE_ALERTS = "Active Alerts"
var HEALTH_REDIS = "Redis status"
var COLLECT_KARGO = "Collect Kargo"
var SET_DEBUG_LEVEL = "Set Debug Level"
var FLUSH_REDIS = "Flush Redis"
var PLACEHOLDER = "Placeholder"

func createApplication() (app *tview.Application) {
	app = tview.NewApplication()
	pages := tview.NewPages()
	infoUI := createInfoPanel(app)
	logPanel := createTextViewPanel(app, "Output Terminal")
	logPanel.SetBorder(true)
	logPanel.SetDynamicColors(true)
	logPanel.SetTextColor(tcell.ColorWhite)

	//do not print date and time
	log.SetFlags(0)
	log.SetOutput(logPanel)
	log.Println("Welcome to HealthCtl")
	log.Println(" [green]✔[-] Version: v1.0")
	log.Println(" [green]✔[-] This is a tool to run sanity checks on k8s clusters and NFs in K8s clusters")
	log.Println(" [green]✔[-] Check Alerts, SMF status, UPF Status, Redis Status, Collect Kargo, Set Debug levels and Flush Redis.")
	log.Println(" [green]✔[-] Use shortcuts to run tests, stop tests, open reports, view alerts and run Popeye.")
	log.Println(" [green]✔[-] Use ctrl+r to run tests, ctrl+s to stop tests, ctrl+o to open reports, a to view alerts and ctrl+p to run Popeye.")
	log.Println(" [green]✔[-] Use arrow keys to navigate and enter to select.")
	log.Println(" [green]✔[-] Use esc to go back to main menu.")
	log.Println(" [green]✔[-] Use q to quit the application.")
	log.Println(" [green]✔[-] Use tab to navigate between tools and output terminal.")
	log.Println(" [green]✔[-] Use mouse to click the buttons in tools.")

	var CreateNewButton func(label string, handler func()) *tview.Button
	CreateNewButton = func(label string, handler func()) *tview.Button {
		button := tview.NewButton(label)
		button.SetBackgroundColor(tcell.ColorGrey)
		button.SetLabelColor(tcell.ColorBlack)
		button.SetBorderColor(tcell.ColorBlack)
		button.SetBorder(false)
		button.SetSelectedFunc(handler)
		button.SetBackgroundColorActivated(tcell.ColorBlack)
		return button
	}

	afn_tools := tview.NewFlex()
	afn_tools.SetDirection(tview.FlexRow)
	afn_tools.SetBorder(true).SetTitle("Tools")
	afn_tools.AddItem(CreateNewButton(HEALTH_K8s, sendCommand(pages, infoUI, HEALTH_K8s)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)

	afn_tools.AddItem(CreateNewButton(HEALTH_INFRA, sendCommand(pages, infoUI, HEALTH_INFRA)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(HEALTH_PAAS, sendCommand(pages, infoUI, HEALTH_PAAS)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(HEALTH_SMF, sendCommand(pages, infoUI, HEALTH_SMF)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(HEALTH_UPF, sendCommand(pages, infoUI, HEALTH_UPF)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(HEALTH_STORAGE, sendCommand(pages, infoUI, HEALTH_STORAGE)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(ACTIVE_ALERTS, Alerts(pages)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(HEALTH_REDIS, RedisStatus(pages)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(COLLECT_KARGO, func() {}), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(SET_DEBUG_LEVEL, SetDebugLevel(pages)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(FLUSH_REDIS, FlushRedis(pages)), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)
	afn_tools.AddItem(CreateNewButton(PLACEHOLDER, func() {}), 0, 1, false)
	afn_tools.AddItem(tview.NewBox(), 1, 0, false)

	layout := createMainLayout(infoUI, logPanel, afn_tools, pages)
	pages.AddPage("main", layout, true, true)

	app.SetRoot(pages, true).EnableMouse(true)
	return app
}

func SetDebugLevel(pages *tview.Pages) func() {
	kc, _ := k8s.NewK8sClient()
	return func() {
		//open a new popup with a form to take input like namespace, podname, container name and debug level
		form := tview.NewForm()
		//form.SetBackgroundColor(tcell.ColorDarkCyan)
		var namespaceSelection, podSelection, containerSelection, levelSelection *tview.DropDown
		namespaceSelection = tview.NewDropDown()
		//namespaceSelection.SetBackgroundColor(tcell.ColorLightCyan)
		namespaceSelection.SetOptions(kc.GetClusterNamespaces(), func(text string, index int) {
			//get all pods in selected namespace
			podSelection.SetOptions(kc.GetPods(text), func(text string, index int) {
				//get all containers in selected pod
				containerSelection.SetOptions(kc.GetContainers(text), func(text string, index int) {
					//Set level selection
					levelSelection.SetOptions([]string{"DEBUG_1", "DEBUG_2", "DEBUG_3"}, nil).SetLabel("Level")
				}).SetLabel("Container")
			}).SetLabel("Pod")
		}).SetLabel("Namespace")

		podSelection = tview.NewDropDown()
		//podSelection.SetBackgroundColor(tcell.ColorLightCyan)
		containerSelection = tview.NewDropDown()
		levelSelection = tview.NewDropDown()

		form.AddFormItem(namespaceSelection)
		form.AddFormItem(podSelection)
		form.AddFormItem(containerSelection)
		form.AddFormItem(levelSelection)

		form.AddButton("Submit", func() {
			_, namespace := form.GetFormItemByLabel("Namespace").(*tview.DropDown).GetCurrentOption()
			_, podName := form.GetFormItemByLabel("Pod").(*tview.DropDown).GetCurrentOption()
			_, containerName := form.GetFormItemByLabel("Container").(*tview.DropDown).GetCurrentOption()
			_, debugLevel := form.GetFormItemByLabel("Level").(*tview.DropDown).GetCurrentOption()
			log.Printf("Setting Debug Level for %s/%s/%s to %s\n", namespace, podName, containerName, debugLevel)
			if kc.SetDebugLevel(namespace, podName, containerName, debugLevel) {
				log.Printf("[green]Debug Level set successfully for Container: %s Pod: %s Namespace: %s[-]\n", containerName, podName, namespace)
			} else {
				log.Printf("[red]Error setting Debug for Container: %s Pod: %s Namespace: %s[-]\n", containerName, podName, namespace)
			}
			pages.SwitchToPage("main")
			pages.RemovePage("modal")
		}).SetButtonsAlign(tview.AlignCenter)
		form.AddButton("Cancel", func() {
			pages.SwitchToPage("main")
			pages.RemovePage("modal")
		}).SetButtonsAlign(tview.AlignCenter)
		form.SetBorder(true).SetTitle("Set Debug Level")
		modal := createModalForm(pages, form, 13, 80)
		pages.AddPage("modal", modal, true, true)
	}
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

func GetSelectedCluster() string {
	kc, _ := k8s.NewK8sClient()
	return kc.GetCurrentCluster()
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
		log.Printf("| %-33s | %-8s | %-24s | %-40s | %-40s\n", "─────────────────────────────────", "────────", "────────────────────────", "────────────────────────────────────────", "────────────────────────────────────────")
	}

	//clear logPanel

	equalFormatter()
	log.Printf("| %-33s | %-8s | %-24s | %-40s | %-40s\n", centerText("Alertname", 33), centerText("Severity", 8), centerText("Starts At", 24), centerText("Pod Name", 40), centerText("Summary", 40))
	equalFormatter()

	for _, alert := range alertList {
		if alert.Severity == "critical" {
			alert.Severity = "[red]" + alert.Severity + "[-:-]"
		} else if alert.Severity == "major" {
			alert.Severity = "[yellow]" + alert.Severity + "[-:-]   "
		} else {
			alert.Severity = "[green]" + alert.Severity + "[-:-] "
		}

		log.Printf("| %-33s | %-8s | %-24s | %-40s | %-40s\n", alert.AlertName, alert.Severity, alert.StartsAt, alert.PodName, alert.Summary)
		equalFormatter()
	}
	log.Printf("| %-33s | %-8s | %-24s | %s | %-40s\n", "", "", "", centerText("Total Alerts", 40), strconv.Itoa(len(alertList)))
	equalFormatter()
}
func createMetadataPanel(infoUI *testInfoUI) *tview.Table {
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
	return metadata
}

func createMainLayout(infoUI *testInfoUI, output tview.Primitive, afn_tools *tview.Flex, pages *tview.Pages) (layout *tview.Flex) {
	///// Main Layout /////
	metadata := createMetadataPanel(infoUI)

	kc, _ := k8s.NewK8sClient()
	config := k8s.GetClustersFromKubeConfig()
	clusters := []string{}
	for index, _ := range config.Clusters {
		clusters = append(clusters, index)
	}
	handler := func(text string, index int) {
		kc.SetContext(config, text)
		infoUI.context.SetText(config.CurrentContext)
		infoUI.cluster.SetText(text)
		nodes := kc.GetClusterNodes()
		infoUI.nodes.SetText(fmt.Sprintf("Master: %d, Worker: %d", nodes[0], nodes[1]))
		infoUI.apiserver.SetText(config.Clusters[text].Server)
		pages.SwitchToPage("main")
	}

	form := tview.NewForm()
	cluster := tview.NewDropDown()
	cluster.SetOptions(clusters, handler).SetCurrentOption(0).SetFieldWidth(30).SetLabel("Cluster")
	form.AddFormItem(cluster).SetBorder(true).SetTitle("Cluster Selection")

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
	banner.SetBorder(true)
	for i := 0; i < 7; i++ {
		banner.SetCell(i+1, 0, tview.NewTableCell(Logo[i]).SetTextColor(tcell.ColorYellow))
		banner.GetCell(i+1, 0).SetAlign(tview.AlignRight)
	}

	mainLayout := tview.NewFlex()
	mainLayout.SetDirection(tview.FlexColumn).
		// AddItem(clusterList, 20, 0, true).
		// AddItem(commandList, 20, 0, true).
		AddItem(afn_tools, 30, 0, false).
		AddItem(output, 0, 30, false)
	mainLayout.SetBackgroundColor(tcell.ColorGrey)

	info := tview.NewTextView()
	info.SetBorder(true)
	info.SetText("HealthCtl v1.0 - © Microsoft 2024")
	info.SetTextAlign(tview.AlignCenter)

	header := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(form, 0, 1, false).
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
	layout.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(0).(*tview.Flex).GetItem(1).(*tview.TextView).Clear()
}

func centerText(text string, width int) string {
	// Calculate padding on both sides
	padding := (width - len(text)) / 2
	if padding > 0 {
		// If padding is positive, center the text
		return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
	}
	// If text is longer than width, just return it as is
	return text
}

func runTests(selectedCommand string) {
	kc, _ := k8s.NewK8sClient()
	rl := []models.ResourceCheck{}
	switch selectedCommand {
	case HEALTH_K8s:
		rl = testsuite.CheckK8s(kc.Client)
		break
	case HEALTH_INFRA:
		rl = testsuite.CheckINFRA(kc.Client)
		break
	case HEALTH_PAAS:
		rl = testsuite.CheckPAAS(kc.Client)
		break
	case HEALTH_SMF:
		rl = testsuite.CheckSMF(kc.Client)
		break
	case HEALTH_UPF:
		rl = testsuite.CheckUPF(kc.Client)
		break
	case HEALTH_STORAGE:
		rl = testsuite.CheckStorage(kc.Client)
		break
	default:
		log.Printf("Please select a test to run")
	}

	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
	log.Printf("| %s | %s | %s  |\n", centerText("No.", 5), centerText("Test Summary", 150), centerText("Result", 7))
	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")

	status := ""
	for index, resc := range rl {
		if resc.Status {
			status = "[:green::]PASS[:-::]"
		} else {
			status = "[:red::]FAIL[:-::]"
		}
		log.Printf("| %s | %-150s | %-7s %s|\n", centerText(strconv.Itoa(index+1), 5), resc.Details, status, "   ")
		log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
	}
	log.Printf("| %-5s | %s | %-7s |\n", "", centerText("Total Tests", 150), strconv.Itoa(len(rl)))
	log.Printf("| %-5s | %-150s | %-7s |\n", "─────", "──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────", "──────")
}

func sendCommand(pages *tview.Pages, infoUI *testInfoUI, selectedCommand string) func() {
	return func() {
		startFunc := func(selectedCommand string) {
			stop(infoUI)()
			pages.SwitchToPage("main")
			clearLogPanel(pages)
			runTests(selectedCommand)
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
		form.SetBackgroundColor(tcell.ColorDarkSlateGray)
		form.AddButton("Start", func() {
			startFunc(selectedCommand)
		})
		form.AddButton("Cancel", cancelFunc)
		form.SetCancelFunc(cancelFunc)
		form.SetButtonsAlign(tview.AlignCenter)

		form.SetBorder(true).SetTitle("Confirmation")
		form.AddTextView(fmt.Sprintf("Executing %s command on %s cluster", selectedCommand, GetSelectedCluster()), "", 0, 1, false, false)

		modal := createModalForm(pages, form, 13, 80)

		pages.AddPage("modal", modal, true, true)

	}
}

func createInfoPanel(app *tview.Application) (infoUI *testInfoUI) {
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
	// kc, _ := k8s.NewK8sClient()
	// fmt.Println(kc.GetResourceUsageReport())
	app := createApplication()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
