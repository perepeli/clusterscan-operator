# ClusterScan Operator

## Overview

The ClusterScan Operator is a custom Kubernetes controller built using Kubebuilder, this operator introduces a `ClusterScan` custom resource that encapsulates Jobs and CronJobs, ensuring controlled execution and descriptive status reporting. The operator reconciles `ClusterScan` resources by creating and managing the lifecycle of these encapsulated workloads, and it reflects their statuses within the `ClusterScan` resource.
### Example `ClusterScan` CRD Definition

Here is the example CRD definition that includes the `ClusterScanSpec` and `ClusterScanStatus`:

```go
type ClusterScanSpec struct {
    Schedule    string          `json:"schedule,omitempty"`
    JobTemplate batchv1.JobSpec `json:"jobTemplate"`
}

type ClusterScanStatus struct {
    LastScheduleTime *metav1.Time             `json:"lastScheduleTime,omitempty"`
    Active           []corev1.ObjectReference `json:"active,omitempty"`
    Conditions       []metav1.Condition       `json:"conditions,omitempty"`
    UnifiedStatus    string                   `json:"unifiedStatus,omitempty"`
    Message          string                   `json:"message,omitempty"`
}
```
## Description

The ClusterScan Operator was created to address specific requirements: to implement a Kubernetes controller that manages `ClusterScan` resources, which act as interfaces for arbitrary Jobs and CronJobs. The operator supports both one-off and recurring executions, leveraging Kubernetes Jobs and CronJobs based on the provided schedule. Key features include encapsulation of Jobs and CronJobs, enhanced security through enforced policies, and detailed status updates that provide visibility into the execution states of the encapsulated workloads. By using ClusterScan, administrators benefit from improved security measures, such as running jobs with restricted privileges and limiting network access, which are not easily achieved with plain Jobs or CronJobs. 

(Work is still in progress on security research part...)



### Status Mapping Table

| Kubernetes Resource | Resource State               | ClusterScan `UnifiedStatus` | ClusterScan `Message`                        |
|---------------------|------------------------------|-----------------------------|----------------------------------------------|
| **Job**             | Active (running)             | Running                     | Job is currently running.                    |
|                     | Succeeded                    | Succeeded                   | Job has completed successfully.              |
|                     | Failed                       | Failed                      | Job has failed.                              |
|                     | Pending                      | Pending                     | Job is pending.                              |
| **CronJob**         | Active (job running)         | Running                     | CronJob has active jobs currently running.   |
|                     | Suspended                    | Suspended                   | CronJob is suspended.                        |
|                     | Scheduled (next run pending) | Scheduled                   | CronJob is scheduled and awaiting next run.  |
|                     | Inactive (no active job)     | Inactive                    | CronJob is inactive.                         |

## Reconcile loop diagram
![Reconcile Diagram](reconcile_diagram.jpg)


## Getting Started

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/clusterscan-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/clusterscan-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/clusterscan-operator:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/clusterscan-operator/<tag or branch>/dist/install.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

