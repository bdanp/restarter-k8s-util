# Kubernetes Resource Restart Tool

This tool restarts Kubernetes resources (Deployments, StatefulSets, and DaemonSets) across all namespaces based on a name filter.

## Prerequisites

- Go 1.18 or later
- Access to a Kubernetes cluster
- Valid kubeconfig file

## Setup and Running the Tool

1. Ensure you have Go installed

    `go version`


2. Clone the repo and change to the repo folder

    `git clone git@github.com:bdanp/restarter-k8s-util.git`


3. Download and sync go modules dependencies

    `go mod tidy`

4. set the environment variables [OPTIONAL] ***provided values are default values***

    `export FILTER_NAME=database`

    `export KUBECONFIG=$HOME/.kube/config`

5. Set the environment variable that exclude namepsace(s) for scanning and rollout restarting the deployments. Default namepsace such as default and kube-system are excluded even the variable is not set. 

    `export EXCLUDE_K8S_NS="kube-system,kube-public,kube-node-lease,prod`

5. Run the script

    `go run main.go`
