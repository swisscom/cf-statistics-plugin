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
package helper

import (
	"bytes"
	"fmt"
	"os/exec"
)

func CallCommandHelp(command string, errorText string) {
	var out bytes.Buffer

	cmd := exec.Command("cf", command, "-h")
	cmd.Stdout = &out
	cmd.Run()

	fmt.Println("\nFAILED")
	fmt.Println(errorText)
	fmt.Println("")
	fmt.Println(out.String())
}
