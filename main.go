package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	corev1 "k8s.io/api/core/v1"
)

func main() {
	filterName := os.Getenv("FILTER_NAME")
	if filterName == "" {
		filterName = "database"
	}

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}

       excludedNamespaces := strings.Split(os.Getenv("EXCLUDE_K8S_NS"), ",")
        if len(excludedNamespaces) == 1 && excludedNamespaces[0] == "" {
                excludedNamespaces = []string{"kube-system", "default"}
        }


	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("error loading Kubernetes configuration :: %v\n", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error creating Kubernetes clientset :: %v\n", err)
		return
	}

	fmt.Println("connecting cluster...")
	ctx := context.Background()

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error listing namespaces :: %v\n", err)
		return
	}

	var filteredNamespaces []corev1.Namespace
	for _, ns := range namespaces.Items {
		if !contains(excludedNamespaces, ns.Name) {
			filteredNamespaces = append(filteredNamespaces, ns)
		}
	}

	fmt.Printf("found %d namespaces\n", len(filteredNamespaces))
	for _, ns := range filteredNamespaces {
		fmt.Printf("searching namespace %s\n", ns.Name)
		deployments, err := clientset.AppsV1().Deployments(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("error listing deployments :: %v\n", err)
			continue
		}

		for _, deployment := range deployments.Items {
			if strings.Contains(strings.ToLower(deployment.Name), filterName) {
				_, err := clientset.AppsV1().Deployments(ns.Name).
					Patch(ctx, deployment.Name, types.StrategicMergePatchType, []byte(GetPatchUpdateAnnotationSpec()), metav1.PatchOptions{})
				if err != nil {
					fmt.Printf("error restarting deployment %s :: %v\n", deployment.Name, err)
					continue
				}
				fmt.Printf("deployment restarted successfully for %s\n", deployment.Name)
			}
		}

		statefulsets, err := clientset.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("error listing statefulsets :: %v\n", err)
			continue
		}

		for _, statefulset := range statefulsets.Items {
			if strings.Contains(strings.ToLower(statefulset.Name), filterName) {
				_, err := clientset.AppsV1().StatefulSets(ns.Name).
					Patch(ctx, statefulset.Name, types.StrategicMergePatchType, []byte(GetPatchUpdateAnnotationSpec()), metav1.PatchOptions{})
				if err != nil {
					fmt.Printf("error restarting statefulset %s :: %v\n", statefulset.Name, err)
					continue
				}
				fmt.Printf("statefulset restarted successfully for %s\n", statefulset.Name)
			}
		}

		daemonsets, err := clientset.AppsV1().DaemonSets(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("error listing daemonsets :: %v\n", err)
			continue
		}

		// Check each daemonset and update rollout annotation
		for _, daemonset := range daemonsets.Items {
			if strings.Contains(strings.ToLower(daemonset.Name), filterName) {
				_, err := clientset.AppsV1().DaemonSets(ns.Name).Patch(ctx, daemonset.Name, types.StrategicMergePatchType, []byte(GetPatchUpdateAnnotationSpec()), metav1.PatchOptions{})
				if err != nil {
					fmt.Printf("error restarting daemonset %s:: %v\n", daemonset.Name, err)
				}
				fmt.Printf("daemonset restarted successfully for %s\n", daemonset.Name)
			}
		}
	}
	fmt.Printf("gracefull restart process completed for pods resource containing keyword %s\n", filterName)
}

func GetPatchUpdateAnnotationSpec() string {
	return fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format(time.RFC3339))
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
