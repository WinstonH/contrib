# Horizontal Self Scaler Sidecar container

This container image watches over the number of schedulable nodes in the cluster and resizes 
the number of replicas in its parent object. Works only for pods which are children of RCs or RSs objects.

Usage of pod_nanny:
    --configFile <params file>
```

# Implementation Details

The code in this module is a Kubernetes Golang API client that learns its own namespace/pod name from 
environment variables populated by downward API field references. Using the default service account credentials
available to Golang clients running inside pods, it connects to the API server and polls for the number of nodes
and cores in the cluster.
The scaling parameters and data points are provided via a ConfigMap to the autoscaler and it refreshes its
parameters table every poll interval to be up to date to the latest desired scaling parameters.

## Calculation of number of replicas

The desired number of replicas is computed by lookup up the number of cores using the step ladder function.
The step ladder function uses the datapoints from the configmap.
This may be later extended to more complex interpolation or linear/exponential scaling schemes
but it currently supports (and defaults to) to mode=step only.

## Derivation of scaling controller object

Using the created-by annotation  (```kubernetes.io/created-by```) which provides a SerializedReference to the parent
object, it is possible to find the parent ReplicationController or ReplicaSet that creates and owns the pod. All
other Kinds of ObjectReferences are not supported at this time.
For ReplicaSets, there may be yet another level of ownership (Deployments), which will be similarly accessed via
their created-by annotations.

Once the top most object owning and controlling the scaling of the pods is derived, the number of replicas is updated
to the desired number, if it is not already the same.

# Central Configmap controlling all sidecar replicas

There will be an autoscaler sidecar container running in every replica - these are all stateless and fetch parameters
from the ConfigMap and the number of nodes every poll interval.
Thus, they all work towards the same scale goals and should converge quickly.

## Example rc file

This [example-rc.yaml](example-rc.yaml) is an example Replication Controller where the nannies in all pods watch and resizes the RC replicas
