// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package commands

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/shirou/gopsutil/v3/mem"
	netutil "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func Init_Monitor_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var cpuDataSize int
	var cpuZoom float64
	var filesDataSize int
	var filesZoom float64
	var interval int
	var memDataSize int
	var memZoom float64
	var netDataSize int
	var netKind string
	var netZoom float64

	var monitorCmd = &cobra.Command{
		Use:     "monitor [pid or name]",
		Aliases: []string{"mon"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Monitor process",
		Long:    `Monitors a process by PID or its name.`,
		Run: func(cmd *cobra.Command, args []string) {
			processes, err := process.Processes()
			utils.CheckForError(err)

			pidOrName := strings.TrimSpace(
				strings.ToLower(args[0]),
			)

			var processToMonitor *process.Process

			i64Val, err := strconv.ParseInt(pidOrName, 10, 32)
			if err == nil {
				// valid number, search by name
				pid := int32(i64Val)

				for _, p := range processes {
					if p.Pid == pid {
						processToMonitor = p
						break
					}
				}
			} else {
				matchingProcesses := []*process.Process{}

				for _, p := range processes {
					name, err := p.Name()
					if err != nil {
						continue
					}

					lowerName := strings.TrimSpace(
						strings.ToLower(name),
					)

					shouldAddProcess := true

					for _, part := range args {
						lowerPart := strings.ToLower(part)
						if !strings.Contains(lowerName, lowerPart) {
							shouldAddProcess = false
							break
						}
					}

					if shouldAddProcess {
						matchingProcesses = append(matchingProcesses, p)
					}
				}

				matchingProcessesCount := len(matchingProcesses)
				if matchingProcessesCount == 1 {
					processToMonitor = matchingProcesses[0]
				} else if matchingProcessesCount > 1 {
					utils.CloseWithError(fmt.Errorf("found %v matching process", matchingProcessesCount))
				}
			}

			if processToMonitor == nil {
				utils.CloseWithError(fmt.Errorf("process %v not found", pidOrName))
			}

			if err := ui.Init(); err != nil {
				log.Fatalf("failed to initialize termui: %v", err)
			}
			defer ui.Close()

			// data
			cpuData := []float64{0}
			filesData := []float64{0}
			memData := []float64{0}
			netData := []float64{0}

			// CPU diagram
			slCpu := widgets.NewSparkline()
			slCpu.Data = cpuData
			slCpu.MaxVal = 100 / cpuZoom

			// memory diagram
			slMem := widgets.NewSparkline()
			slMem.Data = memData

			// network diagram
			slNet := widgets.NewSparkline()
			slNet.Data = netData
			slNet.MaxVal = 100 / netZoom

			// files diagram
			slFiles := widgets.NewSparkline()
			slFiles.Data = filesData

			vMem, err := mem.VirtualMemory()
			if err == nil {
				slMem.MaxVal = float64(vMem.Total) / memZoom
			}

			var rLimit syscall.Rlimit
			err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
			if err == nil {
				if rLimit.Cur != 0 {
					slFiles.MaxVal = float64(rLimit.Cur) / filesZoom
				}
			}

			rerender := func() {
				currentCpu := cpuData[0]
				currentFiles := filesData[0]
				currentMem := memData[0]
				currentNet := netData[0]

				processName, _ := processToMonitor.Name()
				processPid := processToMonitor.Pid

				termWidth, termHeight := ui.TerminalDimensions()

				grid := ui.NewGrid()
				grid.SetRect(0, 3, termWidth, termHeight)

				pTitle := widgets.NewParagraph()
				pTitle.Text = fmt.Sprintf("%v (%v)", processName, processPid)
				pTitle.SetRect(0, 0, termWidth, 3)
				pTitle.Border = true

				totalGridHeight := termHeight - 3
				gridRowHeights := []int{totalGridHeight / 2}
				gridRowHeights = append(gridRowHeights, totalGridHeight-gridRowHeights[0])

				// create a sparkline group from CPU widgets
				slgCpu := widgets.NewSparklineGroup(slCpu)
				slgCpu.Title = fmt.Sprintf(
					"CPU %.2f%% (%.1fx)",
					currentCpu, cpuZoom,
				)
				slgCpu.SetRect(0, 0, termWidth, gridRowHeights[0])

				// create a sparkline group from memory widgets
				slgMem := widgets.NewSparklineGroup(slMem)
				slgMem.Title = fmt.Sprintf(
					"MEM %vMB / %vMB (%.1fx)",
					fmt.Sprintf("%.2f", currentMem/1024.0/1024.0),
					fmt.Sprintf("%.2f", float64(vMem.Total)/1024.0/1024.0),
					memZoom,
				)
				slgMem.SetRect(0, 0, termWidth, gridRowHeights[0])

				// create a sparkline group from net widgets
				slgNet := widgets.NewSparklineGroup(slNet)
				slgNet.Title = fmt.Sprintf(
					"Net %.0f (%.1fx)",
					currentNet, netZoom,
				)
				slgNet.SetRect(0, 0, termWidth, gridRowHeights[1])

				// create a sparkline group from files widgets
				slgFiles := widgets.NewSparklineGroup(slFiles)
				slgFiles.Title = fmt.Sprintf(
					"Files %v / %v (%.1fx)",
					currentFiles, rLimit.Cur,
					filesZoom,
				)
				slgFiles.SetRect(0, 0, termWidth, gridRowHeights[1])

				// create grid ...
				grid.Set(
					// with 2 rows
					ui.NewRow(1.0/2,
						// and 2 columns
						// with 50%/50% size
						ui.NewCol(1.0/2, slgCpu),
						ui.NewCol(1.0/2, slgMem),
					),
					ui.NewRow(1.0/2,
						ui.NewCol(1.0/2, slgNet),
						ui.NewCol(1.0/2, slgFiles),
					),
				)

				// render whole UI
				ui.Render(pTitle, grid)
			}

			shouldRun := true
			go func() {
				for shouldRun {
					// wait before continue
					time.Sleep(time.Duration(interval) * time.Millisecond)

					// memory usage
					memInfo, err := processToMonitor.MemoryInfo()
					if err == nil {
						memData = append([]float64{float64(memInfo.RSS)}, memData...)
					} else {
						memData = append([]float64{-1}, memData...)
					}
					memData = utils.EnsureMaxSliceLength(memData, memDataSize)

					// CPU usage
					cpuPercent, err := processToMonitor.CPUPercent()
					if err == nil {
						cpuData = append([]float64{cpuPercent}, cpuData...)
					} else {
						cpuData = append([]float64{-1}, cpuData...)
					}
					cpuData = utils.EnsureMaxSliceLength(cpuData, cpuDataSize)

					// network usage
					netConnections, err := netutil.Connections("all")
					if err == nil {
						netConnectionCount := len(netConnections)

						netData = append([]float64{float64(netConnectionCount)}, netData...)
					} else {
						netData = append([]float64{-1}, netData...)
					}
					netData = utils.EnsureMaxSliceLength(netData, netDataSize)

					// open files
					numberOfOpenFiles, err := utils.GetNumberOfOpenFilesByPid(processToMonitor.Pid)
					if err == nil {
						filesData = append([]float64{float64(numberOfOpenFiles)}, filesData...)
					} else {
						filesData = append([]float64{-1}, filesData...)
					}
					filesData = utils.EnsureMaxSliceLength(filesData, filesDataSize)

					// update data ...
					utils.UpdateUsageSparkline(slMem, memData)
					utils.UpdateUsageSparkline(slCpu, cpuData)
					utils.UpdateUsageSparkline(slNet, netData)
					utils.UpdateUsageSparkline(slFiles, filesData)

					// .. before rerender
					rerender()
				}
			}()

			rerender() // initial rendering

			uiEvents := ui.PollEvents()
			for {
				e := <-uiEvents
				switch e.ID {
				case "q", "<C-c>":
					// CTRL + C
					shouldRun = false
					return
				}
			}
		},
	}

	monitorCmd.Flags().IntVarP(&cpuDataSize, "cpu-data-size", "", 512, "custom size of maximum data items for CPU sparkline")
	monitorCmd.Flags().Float64VarP(&cpuZoom, "cpu-zoom", "", 1.0, "zoom factor for CPU sparkline")
	monitorCmd.Flags().IntVarP(&filesDataSize, "files-data-size", "", 512, "custom size of maximum data items for files sparkline")
	monitorCmd.Flags().Float64VarP(&filesZoom, "files-zoom", "", 1.0, "zoom factor for files sparkline")
	monitorCmd.Flags().IntVarP(&interval, "interval", "", 500, "time in milliseconds for the update interval")
	monitorCmd.Flags().IntVarP(&memDataSize, "mem-data-size", "", 512, "custom size of maximum data items for mem sparkline")
	monitorCmd.Flags().Float64VarP(&memZoom, "mem-zoom", "", 1.0, "zoom factor for mem sparkline")
	monitorCmd.Flags().IntVarP(&netDataSize, "net-data-size", "", 512, "custom size of maximum data items for net sparkline")
	monitorCmd.Flags().StringVarP(&netKind, "net-kind", "", "all", "zoom factor for net sparkline")
	monitorCmd.Flags().Float64VarP(&netZoom, "net-zoom", "", 1.0, "zoom factor for net sparkline")

	parentCmd.AddCommand(
		monitorCmd,
	)
}
