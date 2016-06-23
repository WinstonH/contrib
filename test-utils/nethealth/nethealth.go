/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	job    string
	jobId  int
	tmpDir string
)

func init() {

	flag.StringVar(&job, "job", "kubernetes-e2e-gce", "Jenkins job name")
	flag.IntVar(&jobId, "id", 18995, "Job Id to look at")
	flag.StringVar(&tmpDir, "tmpdir", "", "Temporary directory for GCS storage")
}

const (
	gsutilPattern = "gs://kubernetes-jenkins/logs/%s/%d/"
)

func cleanup(tmpDir string) {
	log.Println("Wiping out temp directory", tmpDir)
	err := os.RemoveAll(tmpDir)
	if err != nil {
		log.Fatalf("Error cleaning up temp directory %s", tmpDir)
	}
}

func main() {

	flag.Parse()
	fmt.Printf("Analyzing nethealth reports for job %s\n", job)

	var err error
	tmpDir, err = ioutil.TempDir(tmpDir, "jenkins")
	if err != nil {
		log.Fatalf("Failure creating temp directory (%s)", err)
	}

	// Rsync all the files to the temporary directory next
	bucket := fmt.Sprintf(gsutilPattern, job, jobId)
	log.Printf("Downloading logs from %s to %s", bucket, tmpDir)
	cmd := exec.Command("gsutil", "-m", "rsync", "-r", bucket, tmpDir)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error (%s) downloading files from %s to temp dir %s", err, bucket, tmpDir)
	}

	defer cleanup(tmpDir)

	//log.Printf("Job %s status was %d\n", jobId, status)
}
