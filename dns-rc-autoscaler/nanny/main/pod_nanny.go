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

package main

import (
	"errors"
	"log"
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"k8s.io/contrib/dns-rc-autoscaler/nanny"
)

var (
	// Flags to identify the container to nanny.
	// These flags allow the container to be a sidecar or in an independent pod that watches another pod.
	podNamespace      = flag.String("namespace", os.Getenv("MY_POD_NAMESPACE"), "The namespace of the pod. If unspecified, it will use the environment variable MY_POD_NAMESPACE")
	podName           = flag.String("pod", os.Getenv("MY_POD_NAME"), "The name of the pod to watch. If unspecified, it will use the environment variable MY_POD_NAME")
	configMap         = flag.String("configmap", "", "ConfigMap containing our scaling parameters")
	verbose           = flag.Bool("verbose", false, "Turn on verbose logging to stdout")
	pollPeriodSeconds = flag.Int("poll-period-seconds", 10, "The time, in seconds, to poll the dependent container.")
)

func sanityCheckParametersAndEnvironment() error {
	var errorsFound bool
	if *configMap == "" {
		errorsFound = true
		log.Printf("-configmap parameter cannot be empty\n")
	}
	if *pollPeriodSeconds < 1 {
		errorsFound = true
		log.Printf("-poll-period-seconds cannot be less than 1\n")
	}
	// Log all sanity check errors before returning a single error string
	if errorsFound {
		return errors.New("Failed to validate all input parameters")
	}
	return nil
}

func main() {
	// First log our starting config, and then set up.
	log.Printf("Invoked by %v\n", os.Args)
	flag.Parse()
	// Perform further validation of flags.
	if err := sanityCheckParametersAndEnvironment(); err != nil {
		log.Fatal(err)
	}
	k8s := nanny.NewKubernetesClient(*podNamespace, *podName)
	log.Printf("Looking for parent/owner of pod %s/%s\n", *podNamespace, *podName)
	rc, rs, deployment, err := k8s.GetParents(*podNamespace, *podName)
	// We cannot proceed if this pod does not have a parent object that is an RC, RS or Deployment
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Pod: %s/%s ownership references - RC: %s, RS: %s, Deployment: %s\n", rc, rs, deployment, *podNamespace, *podName)
	scaler := nanny.Scaler{ConfigFile: *configFile, Verbose: *verbose}
	var pollPeriod time.Duration = pollPeriodSeconds * time.Second
	// Begin nannying.
	nanny.PollAPIServer(k8s, scaler, pollPeriod, *configMap, *verbose)
}
