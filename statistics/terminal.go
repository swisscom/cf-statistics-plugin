/*
   Copyright 2015 Swisscom (Schweiz) AG

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package statistics

import (
	"fmt"
	"math"
	"time"

	ui "github.com/gizak/termui"
	"github.com/pivotal-golang/bytefmt"
)

var (
	term   *TerminalUI
	colors = []ui.Attribute{
		ui.ColorGreen,
		ui.ColorMagenta,
		ui.ColorRed,
		ui.ColorBlue,
		ui.ColorYellow,
		ui.ColorCyan,
	}
)

type TerminalUI struct {
	Body    *ui.Grid
	Usage   *ui.Par
	Summary *ui.Par
	CPU     *ui.Sparklines
	Memory  []*ui.Gauge
	Disk    *ui.BarChart
}

func newTerminalUI(appName string) error {
	if err := ui.Init(); err != nil {
		return err
	}
	ui.UseTheme("helloworld")

	// usage text
	usageText := fmt.Sprintf(`Show live statistics for [%s]

:Press 'q' or 'ctrl-c' to exit
:Press 'PageUp' to increase app instances
:Press 'PageDown' to decrease app instances`, appName)
	usage := ui.NewPar(usageText)
	usage.Height = 12
	usage.TextFgColor = ui.ColorWhite
	usage.Border.Label = "Usage"
	usage.Border.FgColor = ui.ColorCyan

	// summary text
	summary := ui.NewPar("")
	summary.Height = 12
	summary.TextFgColor = ui.ColorRed
	summary.Border.Label = "Summary"
	summary.Border.FgColor = ui.ColorCyan

	// cpu sparklines
	data := [400]int{}
	line := ui.NewSparkline()
	line.Data = data[:]
	line.Height = 4
	line.LineColor = colors[0]
	cpu := ui.NewSparklines(line)
	cpu.Height = 7
	cpu.Border.Label = "CPU Usage"

	// memory gauges
	mem := make([]*ui.Gauge, 1)
	for i := range mem {
		mem[i] = ui.NewGauge()
		mem[i].Percent = 0
		mem[i].Height = 5
	}

	// disk bars
	disk := ui.NewBarChart()
	disk.Border.Label = "Disk Usage (in MB)"
	disk.Data = []int{0, 0, 0, 0, 0, 0, 0, 0}
	disk.Height = 12
	disk.BarWidth = 10
	disk.DataLabels = []string{"I: 0", "I: 1", "I: 2", "I: 3", "I: 4", "I: 5", "I: 6", "I: 7"}
	disk.TextColor = ui.ColorWhite
	disk.BarColor = ui.ColorYellow
	disk.NumColor = ui.ColorWhite

	term = &TerminalUI{ui.Body, usage, summary, cpu, mem, disk}
	return nil
}

func (t *TerminalUI) Close() {
	ui.Close()
}

func (t *TerminalUI) Resize() {
	ui.Body.Width = ui.TermWidth()
	ui.Body.Align()
	ui.Render(ui.Body)
}

func (t *TerminalUI) Render() {
	t.Resize()
}

func (t *TerminalUI) AdjustCPU(stats Statistics) {
	// adjust number of sparklines to number of app instances
	if len(stats.Instances) < len(t.CPU.Lines) {
		t.CPU.Lines = t.CPU.Lines[:len(stats.Instances)]
	}
	if len(stats.Instances) > len(t.CPU.Lines) {
		for i := len(t.CPU.Lines); i < len(stats.Instances); i++ {
			// show max 8 instances
			if len(t.CPU.Lines) > 7 {
				break
			}

			// new sparkline
			data := [400]int{}
			line := ui.NewSparkline()
			line.Data = data[:]
			line.LineColor = colors[i%6]

			// add sparkline
			t.CPU.Lines = append(t.CPU.Lines, line)
		}
	}
	t.CPU.Height = len(t.CPU.Lines)*(13-min(len(stats.Instances), 8)) + 2

	// calculate and update data values
	for i, idx := range stats.Instances {
		// show max 8 instances
		if i > 7 {
			break
		}
		cpu := int(math.Ceil(stats.Data[idx].Stats.Usage.CPU * 100 * 100 * 100))
		t.CPU.Lines[i].Data = append(t.CPU.Lines[i].Data[1:], cpu)
		t.CPU.Lines[i].Title = fmt.Sprintf("Instance %d: %.2f%%", i, stats.Data[idx].Stats.Usage.CPU*100.0)
		t.CPU.Lines[i].Height = 12 - min(len(stats.Instances), 8)
	}
}

func (t *TerminalUI) AdjustMemory(stats Statistics) {
	// memory gauges
	mem := make([]*ui.Gauge, len(stats.Instances))
	for i, idx := range stats.Instances {
		// show max 8 instances
		if i > 7 {
			break
		}

		memory := uint64(stats.Data[idx].Stats.Usage.Memory)
		quota := uint64(stats.Data[idx].Stats.MemoryQuota)
		percent := int(math.Ceil((float64(memory) / float64(quota)) * 100.0))
		mem[i] = ui.NewGauge()
		mem[i].Percent = percent
		mem[i].Height = 13 - min(len(stats.Instances), 8)
		mem[i].Border.Label = fmt.Sprintf("Memory - Instance %d: %d%% (%s / %s)",
			i, percent, bytefmt.ByteSize(memory), bytefmt.ByteSize(quota))
		mem[i].Border.FgColor = ui.ColorWhite
		mem[i].Border.LabelFgColor = ui.ColorWhite
		mem[i].BarColor = colors[i%6]
		mem[i].PercentColor = ui.ColorWhite
	}
	t.Memory = mem

	// update layout
	ui.Body.Rows = []*ui.Row{
		ui.NewRow(
			ui.NewCol(3, 0, t.Usage),
			ui.NewCol(3, 0, t.Summary),
			ui.NewCol(6, 0, t.Disk)),
		ui.NewRow(
			ui.NewCol(6, 0, t.CPU),
			t.newMemCol(6, 0, t.Memory)),
	}
}

func (t *TerminalUI) AdjustDisk(stats Statistics) {
	var quota int64
	var data []int
	for i, idx := range stats.Instances {
		// show max 8 instances
		if i > 7 {
			break
		}

		mb := int((stats.Data[idx].Stats.Usage.Disk / 1024) / 1024)
		data = append(data, mb)

		quota = stats.Data[idx].Stats.DiskQuota
	}
	t.Disk.Data = data
	t.Disk.BarWidth = 20 - min(len(stats.Instances), 8)
	t.Disk.Border.Label = fmt.Sprintf("Disk Usage (in MB) - Quota: %s", bytefmt.ByteSize(uint64(quota)))
}

func (t *TerminalUI) AdjustSummary(stats Statistics) {
	var uptime int64
	var cpuUsage float64
	var memUsage, memQuota, diskUsage, diskQuota int64
	for _, idx := range stats.Instances {
		uptime = max(uptime, stats.Data[idx].Stats.Uptime)
		cpuUsage += stats.Data[idx].Stats.Usage.CPU
		memUsage += stats.Data[idx].Stats.Usage.Memory
		diskUsage += stats.Data[idx].Stats.Usage.Disk
		memQuota += stats.Data[idx].Stats.MemoryQuota
		diskQuota += stats.Data[idx].Stats.DiskQuota
	}
	up, _ := time.ParseDuration(fmt.Sprintf("%ds", uptime))

	summaryText := fmt.Sprintf(
		`Instances running: %d

Uptime: %s
Up Since: %s

Total CPU Usage: %.4f%%
Total Memory Usage: %s / %s
Total Disk Usage: %s / %s`,
		len(stats.Instances),
		up.String(),
		time.Now().Add(-1*up).Format("Mon, 02 Jan 2006 15:04 MST"),
		cpuUsage*100.0,
		bytefmt.ByteSize(uint64(memUsage)), bytefmt.ByteSize(uint64(memQuota)),
		bytefmt.ByteSize(uint64(diskUsage)), bytefmt.ByteSize(uint64(diskQuota)))
	t.Summary.Text = summaryText
}

func (t *TerminalUI) newMemCol(span, offset int, mem []*ui.Gauge) *ui.Row {
	// soooo ugly.. :(
	switch len(mem) {
	case 0:
		return ui.NewCol(span, offset, nil)
	case 1:
		return ui.NewCol(span, offset, mem[0])
	case 2:
		return ui.NewCol(span, offset, mem[0], mem[1])
	case 3:
		return ui.NewCol(span, offset, mem[0], mem[1], mem[2])
	case 4:
		return ui.NewCol(span, offset, mem[0], mem[1], mem[2], mem[3])
	case 5:
		return ui.NewCol(span, offset, mem[0], mem[1], mem[2], mem[3], mem[4])
	case 6:
		return ui.NewCol(span, offset, mem[0], mem[1], mem[2], mem[3], mem[4], mem[5])
	case 7:
		return ui.NewCol(span, offset, mem[0], mem[1], mem[2], mem[3], mem[4], mem[5], mem[6])
	default:
		return ui.NewCol(span, offset, mem[0], mem[1], mem[2], mem[3], mem[4], mem[5], mem[6], mem[7])
	}
	return nil
}

func (t *TerminalUI) ScaleApp(appName string, instances int) {
	// scaling text
	scaling := ui.NewPar(fmt.Sprintf("\nSCALING [%s] TO [%d] INSTANCES...\n", appName, instances))
	scaling.Height = 5
	scaling.TextFgColor = ui.ColorYellow
	scaling.Border.Label = "Scale"
	scaling.Border.FgColor = ui.ColorRed
	scaling.Border.LabelFgColor = ui.ColorWhite
	scaling.Border.LabelBgColor = ui.ColorRed

	ui.Body.Rows = []*ui.Row{ui.NewRow(
		ui.NewCol(8, 2, scaling),
	)}

	term.Render()
}

func (t *TerminalUI) UpdateStatistics(stats Statistics) {
	t.AdjustCPU(stats)
	t.AdjustMemory(stats)
	t.AdjustDisk(stats)
	t.AdjustSummary(stats)

	t.Render()
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
