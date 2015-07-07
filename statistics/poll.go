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
	"sort"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

type Statistics struct {
	Instances []string
	Data      map[string]Instance
}
type Instance struct {
	State string `json:"state"`
	Stats struct {
		Name                string   `json:"name"`
		URIs                []string `json:"uris"`
		Host                string   `json:"host"`
		Port                int      `json:"port"`
		Uptime              int64    `json:"uptime"`
		MemoryQuota         int64    `json:"mem_quota"`
		DiskQuota           int64    `json:"disk_quota"`
		FiledescriptorQuota int      `json:"fds_quota"`
		Usage               struct {
			Time   string  `json:"time"`
			CPU    float64 `json:"cpu"`
			Memory int64   `json:"mem"`
			Disk   int64   `json:"disk"`
		} `json:"usage"`
	} `json:"stats"`
}

func pollStats(cliConnection plugin.CliConnection, appGuid string, outChan chan<- Statistics, errChan chan<- error) {
	cmd := []string{"curl", fmt.Sprintf("/v2/apps/%v/stats", appGuid)}

	for {
		// lock mutex, to avoid colliding with possible other cli commands, like for example app scaling
		commandLock.Lock()
		output, err := cliConnection.CliCommandWithoutTerminalOutput(cmd[:]...)
		commandLock.Unlock()
		if err != nil {
			for _, e := range output {
				fmt.Println(e)
			}
			errChan <- err
			break
		}

		var stats Statistics
		if err := json.Unmarshal([]byte(strings.Join(output, "")), &stats.Data); err != nil {
			errChan <- err
			break
		}

		// get instance_index list & sort it
		var instances []string
		for key, _ := range stats.Data {
			instances = append(instances, key)
		}
		sort.StringSlice(instances).Sort()
		stats.Instances = instances

		// send statistics back
		outChan <- stats

		time.Sleep(time.Second * 1)
	}
}
