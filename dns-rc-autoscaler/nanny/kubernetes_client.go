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
	"encoding/json"
	"fmt"
	"log"

	api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"k8s.io/kubernetes/pkg/client/restclient"
)

const replicationControllerKind = "ReplicationController"
const replicaSetKind = "ReplicaSet"
const deploymentKind = "Deployment"
const createdByAnnotationName = "kubernetes.io/created-by"

type kubernetesClient struct {
	namespace  string
	pod        string
	rc         string
	rs         string
	deployment string
	clientset  *client.Clientset
}

func unmarshalCreatedByAnnotation(annotation string) (*api.SerializedReference, error) {
	var annot api.SerializedReference
	err := json.Unmarshal([]byte(annotation), &annot)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse created-by annotation (%s)", err)
	}
	return &annot, nil
}

// Find this pod's owner references, find the RC or RS parent and return its name
func (k *kubernetesClient) getParentRcRs(namespace, podname string) (string, string, error) {
	// This is acquired from the pod's annotation "kubernetes.io/created-by"
	pod, err := k.clientset.CoreClient.Pods(namespace).Get(podname)
	if err != nil {
		return "", "", err
	}
	createdBy, ok := pod.ObjectMeta.Annotations[createdByAnnotationName]
	if !ok || len(createdBy) == 0 {
		// Empty or missing created-by annotation would mean this is a
		// standalone pod and does not have an RC/RS managing it
		return "", "", fmt.Errorf("Pod does not have a created-by annotation")
	}
	annot, err := unmarshalCreatedByAnnotation(createdBy)
	if annot.Reference.Kind == replicationControllerKind {
		return annot.Reference.Name, "", nil
	} else if annot.Reference.Kind == replicaSetKind {
		return "", annot.Reference.Name, nil
	}
	return "", "", fmt.Errorf("This pod %s/%s was not created by a replication controller or replicaset", namespace, podname)
}

func (k *kubernetesClient) getParentDeployment(namespace, rsname string) (string, error) {
	// This is acquired from the RS's annotation "kubernetes.io/created-by"
	rs, err := k.clientset.ExtensionsClient.ReplicaSets(namespace).Get(rsname)
	if err != nil {
		return "", err
	}
	createdBy, ok := rs.ObjectMeta.Annotations[createdByAnnotationName]
	if !ok || len(createdBy) == 0 {
		// Empty or missing created-by annotation would mean this is a
		// standalone replicaset and does not have a deployment managing it
		return "", fmt.Errorf("Replica-Set does not have a created-by annotation")
	}
	annot, err := unmarshalCreatedByAnnotation(createdBy)
	if annot.Reference.Kind == deploymentKind {
		return annot.Reference.Name, nil
	}
	return "", fmt.Errorf("This replicaset %s/%s was not created by a deployment", namespace, rsname)
}

// GetParents Find this pod's owner references, find the RC or RS parents, and deployment if RS
func (k *kubernetesClient) GetParents(namespace, podname string) (rc, rs, deployment string, err error) {
	rc, rs, err = k.getParentRcRs(namespace, podname)
	if err != nil {
		return
	}
	if len(rs) > 0 {
		deployment, _ := k.getParentDeployment(namespace, rs)
	}
	k.rc = rc
	k.rs = rs
	k.deployment = deployment
	return
}

// CountNodes Count schedulable nodes and cores in our cluster
func (k *kubernetesClient) CountNodes() (totalNodes, schedulableNodes, totalCores, schedulableCores int32, err error) {
	opt := api.ListOptions{Watch: false}

	nodes, err := k.clientset.CoreClient.Nodes().List(opt)
	if err != nil {
		log.Println(err)
		return 0, 0, 0, 0, err
	}
	totalNodes = int32(len(nodes.Items))
	var tc resource.Quantity
	var sc resource.Quantity
	for _, node := range nodes.Items {
		tc.Add(node.Status.Capacity[apiv1.ResourceCPU])
		if !node.Spec.Unschedulable {
			schedulableNodes++
			sc.Add(node.Status.Capacity[apiv1.ResourceCPU])
		}
	}

	tcInt64, tcOk := tc.AsInt64()
	scInt64, scOk := sc.AsInt64()
	if !tcOk || !scOk {
		log.Println("Unable to compute integer values of schedulable cores in the cluster")
		return 0, 0, 0, 0, fmt.Errorf("Unable to compute number of cores in cluster")
	}
	return totalNodes, schedulableNodes, int32(tcInt64), int32(scInt64), nil
}

// PodReplicas Get number of replicas configured in the parent RC/Deployment/RS, in that order
func (k *kubernetesClient) PodReplicas() (int32, error) {
	if len(k.rc) > 0 {
		rc, err := k.clientset.CoreClient.ReplicationControllers(k.namespace).Get(k.rc)
		if err != nil {
			return 0, err
		}
		return int32(*rc.Spec.Replicas), nil
	} else if len(k.deployment) > 0 {
		deployment, err := k.clientset.ExtensionsClient.Deployments(k.namespace).Get(k.deployment)
		if err != nil {
			return 0, err
		}
		return int32(*deployment.Spec.Replicas), nil
	}
	rs, err := k.clientset.ExtensionsClient.ReplicaSets(k.namespace).Get(k.rs)
	if err != nil {
		return 0, err
	}
	return int32(*rs.Spec.Replicas), nil
}

// Update the number of replicas in the parent replication controller
func (k *kubernetesClient) UpdateReplicas(replicas int32) error {
	if replicas == 0 {
		log.Fatalf("Cannot update to 0 replicas")
	}
	if len(k.rc) > 0 {
		rc, err := k.clientset.CoreClient.ReplicationControllers(k.namespace).Get(k.rc)
		if err != nil {
			return err
		}
		*rc.Spec.Replicas = replicas
		_, err = k.clientset.CoreClient.ReplicationControllers(k.namespace).Update(rc)
		if err != nil {
			return err
		}
	} else if len(k.deployment) > 0 {
		deployment, err := k.clientset.ExtensionsClient.Deployments(k.namespace).Get(k.deployment)
		if err != nil {
			return err
		}
		*deployment.Spec.Replicas = replicas
		_, err = k.clientset.ExtensionsClient.Deployments(k.namespace).Update(deployment)
		if err != nil {
			return err
		}
	}
	rs, err := k.clientset.ExtensionsClient.ReplicaSets(k.namespace).Get(k.rs)
	if err != nil {
		return err
	}
	*rs.Spec.Replicas = replicas
	_, err = k.clientset.ExtensionsClient.ReplicaSets(k.namespace).Update(rs)
	if err != nil {
		return err
	}

	return nil
}

// NewKubernetesClient gives a KubernetesClient with the given dependencies.
func NewKubernetesClient(namespace, pod string) KubernetesClient {
	config, err := restclient.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := client.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return &kubernetesClient{
		namespace: namespace,
		pod:       pod,
		clientset: clientset,
	}
}
