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
	"fmt"
	"os"
)

type ReplicaParams struct {
	CoresPerReplica int `json:"cores_per_replica"`
}

type CacheParams struct {
	EntriesPerCore int `json:"entries_per_core"`
}

type MemoryParams struct {
	KibPerSvc int `json:"kib_per_svc"`
	MinKib    int `json:"min_kib"`
}

type CpuParams struct {
	RequestsPerCore int `json:"requests_per_core"`
}

type ScalerParams struct {
	Replicas ReplicaParams `json:"replicas"`
	Cache    CacheParams   `json:"cache"`
	Memory   MemoryParams  `json:"memory"`
	Cpu      CpuParams     `json:"cpu"`
}

// ParseScalerParamsFile Parse the scaler params JSON file
func ParseScalerParamsFile(filename string) (params *ScalerParams, err error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Params file %s does not exist", filename)
		}
		return nil, fmt.Errorf("Params file %s is not readable", filename)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return params, nil
}
