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
package main

import (
	"github.com/cloudfoundry/cli/plugin"
	"github.com/swisscom/cf-statistics-plugin/statistics"
)

type StatisticsPlugin struct{}

func main() {
	plugin.Start(&StatisticsPlugin{})
}

func (s *StatisticsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Statistics",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "statistics",
				Alias:    "stats",
				HelpText: "display live metrics/statistics about an app",
				UsageDetails: plugin.Usage{
					Usage: "cf statistics APP_NAME [--debug] [--full]",
					Options: map[string]string{
						"debug": "Prints metrics to stdout in JSON format",
						"full":  "Prints full statistics to stdout if --debug is enabled",
					},
				},
			},
		},
	}
}

func (s *StatisticsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	switch args[0] {
	case "statistics", "stats":
		statistics.Run(cliConnection, args[1:])
	}
}
