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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/plugin"
	ui "github.com/gizak/termui"
	"github.com/swisscom/cf-statistics-plugin/helper"
)

var (
	commandLock = &sync.Mutex{}
)

type Search struct {
	Resources []struct {
		Metadata struct {
			Guid string `json:"guid"`
		} `json:"metadata"`
	} `json:"resources"`
}

func Run(cliConnection plugin.CliConnection, args []string) {
	if len(args) == 0 {
		helper.CallCommandHelp("statistics", "Incorrect usage.\n")
		os.Exit(1)
	}

	var debugOutput, fullOutput bool
	for _, arg := range args[1:] {
		switch arg {
		case "--debug":
			debugOutput = true
		case "--full":
			fullOutput = true
		}
	}

	appName := args[0]
	appGuid := getAppGuid(cliConnection, appName)

	statsChan := make(chan Statistics)
	errChan := make(chan error)
	defer close(statsChan)
	defer close(errChan)

	go pollStats(cliConnection, appGuid, statsChan, errChan)

	// setup ui
	if !debugOutput {
		if err := newTerminalUI(appName); err != nil {
			fmt.Println("\nERROR:", err)
			os.Exit(1)
		}
		defer term.Close()
	}

	// main loop
	var nofInstances int
	for {
		select {

		// ui events
		case e := <-ui.EventCh():
			if e.Type == ui.EventKey {
				if e.Ch == 'q' || e.Key == ui.KeyCtrlC {
					term.Close()
					os.Exit(0)
				} else if e.Key == ui.KeyPgup {
					// ui shows max 8 instances
					if nofInstances < 8 {
						scaleApp(cliConnection, appName, nofInstances+1)
					}
				} else if e.Key == ui.KeyPgdn {
					if nofInstances > 1 {
						scaleApp(cliConnection, appName, nofInstances-1)
					}
				}
			}

			if e.Type == ui.EventResize {
				term.Resize()
			}

			if e.Type == ui.EventError {
				term.Close()
				fmt.Println("\nERROR:", e.Err)
				os.Exit(0)
			}

		case stats := <-statsChan:
			nofInstances = len(stats.Instances)

			// print to stdout if --debug is set
			if debugOutput {
				for _, idx := range stats.Instances {
					var data interface{}

					// print only usage metrics if --full is not set
					data = stats.Data[idx].Stats.Usage
					if fullOutput {
						data = stats.Data[idx]
					}

					output, err := json.Marshal(data)
					if err != nil {
						fmt.Println("\nERROR:", err)
						os.Exit(1)
					}
					fmt.Printf("{\"instance_index\":\"%s\",\"metrics\":%v}\n", idx, string(output))
				}
			} else {
				// render terminal dashboard
				term.UpdateStatistics(stats)
			}

		case err := <-errChan:
			if !debugOutput {
				term.Close()
			}
			fmt.Println("\nERROR:", err)
			os.Exit(1)

		case <-time.After(time.Second * 15):
			if !debugOutput {
				term.Close()
			}
			fmt.Println("\nTIMEOUT")
			fmt.Println("Querying metrics took too long.. Check your connectivity!")
			os.Exit(1)
		}
	}
}

func getAppGuid(cliConnection plugin.CliConnection, appName string) string {
	repo := core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), func(err error) {
		if err != nil {
			fmt.Println("\nERROR:", err)
			os.Exit(1)
		}
	})
	spaceGuid := repo.SpaceFields().Guid

	cmd := []string{"curl", fmt.Sprintf("/v2/spaces/%v/apps?q=name:%v&inline-relations-depth=1", spaceGuid, appName)}
	output, err := cliConnection.CliCommandWithoutTerminalOutput(cmd...)
	if err != nil {
		for _, e := range output {
			fmt.Println(e)
		}
		os.Exit(1)
	}

	search := &Search{}
	if err := json.Unmarshal([]byte(strings.Join(output, "")), &search); err != nil {
		fmt.Println("\nERROR:", err)
		os.Exit(1)
	}

	return search.Resources[0].Metadata.Guid
}

func scaleApp(cliConnection plugin.CliConnection, appName string, instances int) {
	// lock mutex, to avoid colliding with other cli commands
	commandLock.Lock()
	defer commandLock.Unlock()

	term.ScaleApp(appName, instances)

	cmd := []string{"scale", appName, "-i", fmt.Sprintf("%d", instances)}
	if _, err := cliConnection.CliCommandWithoutTerminalOutput(cmd...); err != nil {
		term.Close()
		fmt.Println("\nERROR:", err)
		os.Exit(0)
	}
}
