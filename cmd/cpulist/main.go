/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	cpuset "github.com/openshift-kni/debug-tools/pkg/k8s_imported"
	"github.com/openshift-kni/debug-tools/pkg/procs"
)

func main() {
	var procfsRoot = flag.StringP("procfs", "P", "/proc", "procfs root")
	var cpuList = flag.StringP("cpu-list", "c", "", "cpulist to split")
	var srcFile = flag.StringP("from-file", "f", "", "read the cpulist to split from the given file")
	flag.Parse()

	var cpus cpuset.CPUSet

	if *srcFile != "" {
		var err error
		var data []byte
		if *srcFile == "-" {
			data, err = ioutil.ReadAll(os.Stdin)
		} else {
			data, err = ioutil.ReadFile(*srcFile)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading cpulist from %q: %v\n", *srcFile, err)
			os.Exit(2)
		}
		cpus = parseCPUsOrDie(strings.TrimSpace(string(data)))
	} else if *cpuList != "" {
		cpus = parseCPUsOrDie(*cpuList)
	} else {
		cpus = allowedCPUsOrDie(*procfsRoot)
	}
	printCPUList(cpus)
}

func parseCPUsOrDie(cpuList string) cpuset.CPUSet {
	cpus, err := cpuset.Parse(cpuList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing %q: %v\n", cpuList, err)
		os.Exit(2)
	}
	return cpus
}

func allowedCPUsOrDie(procfsRoot string) cpuset.CPUSet {
	nullLog := log.New(ioutil.Discard, "", 0)
	ph := procs.New(nullLog, procfsRoot)
	info, err := ph.FromPID(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading process status for pid self: %v\n", err)
		os.Exit(4)
	}
	// consolidate all the cpus:
	b := cpuset.NewBuilder()
	for _, tidInfo := range info.TIDs {
		for _, cpuId := range tidInfo.Affinity {
			b.Add(cpuId)
		}
	}
	return b.Result()
}

func printCPUList(cpus cpuset.CPUSet) {
	for _, cpu := range cpus.ToSlice() {
		fmt.Printf("%v\n", cpu)
	}
}
