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

	ctx := context.Background()


	filteredNamespaces,_ := GetFilteredNameSpaces(ctx, clientset,excludedNamespaces)
	fmt.Printf("found %d namespaces from fn\n", len(filteredNamespaces))

	for _, ns := range filteredNamespaces {
		fmt.Printf("searching namespace %s\n", ns.Name)

		restartMatchingDeployments(ctx, clientset, ns.Name, filterName)
		restartMatchingStatefulSets(ctx, clientset, ns.Name, filterName)
		restartMatchingDaemonSets(ctx, clientset, ns.Name, filterName)


	}
	fmt.Printf("gracefull restart process completed for pods resource containing keyword %s\n", filterName)
}

func GetFilteredNameSpaces(ctx context.Context, clientset *kubernetes.Clientset,excludedNamespaces []string) ([]corev1.Namespace,error)  {

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
        if err != nil {
                fmt.Printf("error listing namespaces :: %v\n", err)
                return nil, err
        }

        var filteredNamespaces []corev1.Namespace
        for _, ns := range namespaces.Items {
                if !contains(excludedNamespaces, ns.Name) {
                        filteredNamespaces = append(filteredNamespaces, ns)
                }
        }

        fmt.Printf("found %d namespaces\n", len(filteredNamespaces))
	return filteredNamespaces,nil
}


 func restartMatchingDeployments(ctx context.Context, clientset *kubernetes.Clientset, namespace, filterName string) error {
  deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
  if err != nil {
    return fmt.Errorf("error listing deployments: %w", err)
  }

  for _, deployment := range deployments.Items {
     fmt.Printf("Scanning Deployment %s\n",deployment.Name)
    if strings.Contains(strings.ToLower(deployment.Name), strings.ToLower(filterName)) {
      _, err := clientset.AppsV1().Deployments(namespace).
        Patch(ctx, deployment.Name, types.StrategicMergePatchType, []byte(GetPatchUpdateAnnotationSpec()), metav1.PatchOptions{})
      if err != nil {
        return fmt.Errorf("error restarting deployment %s: %w", deployment.Name, err)
      }
      fmt.Printf("deployment restarted successfully for %s\n", deployment.Name)
    }
  }

  return nil
}


func restartMatchingStatefulSets(ctx context.Context, clientset *kubernetes.Clientset, namespace, filterName string) error {
	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing statefulsets: %w", err)
	}

	for _, statefulset := range statefulsets.Items {
	 	fmt.Printf("Scanning Statefulset %s\n",statefulset.Name)
		if strings.Contains(strings.ToLower(statefulset.Name), strings.ToLower(filterName)) {
			_, err := clientset.AppsV1().StatefulSets(namespace).
				Patch(ctx, statefulset.Name, types.StrategicMergePatchType, []byte(GetPatchUpdateAnnotationSpec()), metav1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("error restarting statefulset %s: %w", statefulset.Name, err)
			}
			fmt.Printf("statefulset restarted successfully for %s\n", statefulset.Name)
		}
	}

	return nil
}


func restartMatchingDaemonSets(ctx context.Context, clientset *kubernetes.Clientset, namespace string, filterName string) error {
    daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("error listing daemonsets: %v", err)
    }

    for _, daemonset := range daemonsets.Items {
	fmt.Printf("Scanning daemonset %s\n",daemonset.Name)
        if strings.Contains(strings.ToLower(daemonset.Name), filterName) {
            _, err := clientset.AppsV1().DaemonSets(namespace).Patch(
                ctx,
                daemonset.Name,
                types.StrategicMergePatchType,
                []byte(GetPatchUpdateAnnotationSpec()),
                metav1.PatchOptions{},
            )
            if err != nil {
                return fmt.Errorf("error restarting daemonset %s: %v", daemonset.Name, err)
            }
            fmt.Printf("daemonset restarted successfully for %s\n", daemonset.Name)
        }
    }

    return nil
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
