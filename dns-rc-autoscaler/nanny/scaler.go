/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package nanny

import (
	"log"
	"sort"
)

// Scaler determines the number of replicas to run
type Scaler struct {
	ConfigMap string
	Verbose   bool
}

func (s Scaler) scaleWithNodesAndCores(numCurrentNodes, schedulableCores int32) int32 {

	newMap, err := ParseScalerParamsFile(s.ConfigFile)
	if err != nil {
		log.Fatalf("Parse failure: The configmap volume file is malformed (%s)\n", err)
	}

	// construct a search ladder from the map
	keys := make([]int, 0, len(newMap))
	for k := range newMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var neededReplicas int32 = 1

	// Walk the search ladder to get the correct number of replicas
	// for the current number of cores
	for _, coreCount := range keys {
		replicas := newMap[int32(coreCount)]
		if int32(coreCount) > schedulableCores {
			break
		}
		neededReplicas = replicas
	}
	// Minimum of two replicas if there are atleast 2 schedulable nodes
	if numCurrentNodes > 1 && neededReplicas < 2 {
		neededReplicas = 2
	}

	return neededReplicas
}
