package state

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AllEthernetesRessources(clientset *kubernetes.Clientset) {
	namespace := corev1.NamespaceDefault
	labelSelector := "project=ethernetes"
	fmt.Printf("Listing all resources in namespace %q with label %q:\n", namespace, labelSelector)

	// List deployments
	ListDeployments(clientset, namespace, labelSelector)

	// List services
	ListServices(clientset, namespace, labelSelector)

	// List pods
	ListPods(clientset, namespace, labelSelector)
}

func ListDeployments(clientset *kubernetes.Clientset, namespace, labelSelector string) {
	fmt.Println("Deployments:")
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	deployList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Printf("Error listing deployments: %v\n", err)
		return
	}
	for _, d := range deployList.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}

func ListServices(clientset *kubernetes.Clientset, namespace, labelSelector string) {
	fmt.Println("Services:")
	servicesClient := clientset.CoreV1().Services(namespace)
	serviceList, err := servicesClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Printf("Error listing services: %v\n", err)
		return
	}
	for _, s := range serviceList.Items {
		fmt.Printf(" * %s\n", s.Name)
	}
}

func ListPods(clientset *kubernetes.Clientset, namespace, labelSelector string) {
	fmt.Println("Pods:")
	podsClient := clientset.CoreV1().Pods(namespace)
	podList, err := podsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		return
	}
	for _, p := range podList.Items {
		fmt.Printf(" * %s\n", p.Name)
	}
}
