package codeserver

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// TODO :
// - Update the the ingress to expose the codeserver
// - Also configure random number for labels completion
// - Return the medata label of the pod to give it to the service
// It is necessary to deploy more than one deployment
func StartDeploymentCodeServer(clientset *kubernetes.Clientset, namespace string, randomNumber int) {

	// Path to deployment YAML file
	yamlFilepath := filepath.Join("./manifests/code-server/code-server-deployment.yaml")

	// Read yaml file
	yamlFile, err := os.ReadFile(yamlFilepath)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return
	}

	// Déserializer le fichier YAML dans un objet Deployment
	// Créer un deserializer pour les objets Kubernetes à partir du schéma de client-go.
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()

	var deployment appsv1.Deployment
	_, _, err = deserializer.Decode(yamlFile, nil, &deployment)
	if err != nil {
		fmt.Printf("Error decoding YAML to Deployment: %s\n", err)
		return
	}

	// Update Deployment object
	deploymentName := fmt.Sprintf("code-server-deployment-%d", randomNumber)
	deployment.ObjectMeta.Name = deploymentName
	deployment.ObjectMeta.Namespace = namespace
	deployment.Labels["number"] = fmt.Sprintf("%d", randomNumber)
	deployment.Spec.Template.ObjectMeta.Labels["number"] = fmt.Sprintf("%d", randomNumber)

	// Create Deployment
	fmt.Printf("Creating deployment %s...\n", deploymentName)
	deploymentsClient := clientset.AppsV1().Deployments(deployment.Namespace)
	result, err := deploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating/updating deployment: %s\n", err)
		return
	}
	fmt.Printf("Deployment %s created/updated.\n", result.GetObjectMeta().GetName())
}

func StartServiceCodeServer(clientset *kubernetes.Clientset, namespace string, randomNumber int) (string, error) {
	// Path to deployment YAML file
	yamlFilepath := filepath.Join("./manifests/code-server/code-server-svc.yaml")

	// Read the YAML file
	yamlFile, err := os.ReadFile(yamlFilepath)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return "", err
	}

	// Deserialize the YAML file into a Service object
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()

	var service apiv1.Service
	_, _, err = deserializer.Decode(yamlFile, nil, &service)
	if err != nil {
		fmt.Printf("Error decoding YAML to Service: %s\n", err)
		return "", err
	}

	serviceName := fmt.Sprintf("code-server-service-%d", randomNumber)
	service.ObjectMeta.Name = serviceName
	service.ObjectMeta.Namespace = namespace
	service.Labels["number"] = fmt.Sprintf("%d", randomNumber)
	service.Spec.Selector["number"] = fmt.Sprintf("%d", randomNumber)

	fmt.Println("Creating service for code-server...")
	result, err := clientset.CoreV1().Services(apiv1.NamespaceDefault).Create(context.TODO(), &service, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating service: %v\n", err)
		return "", err
	}
	fmt.Printf("Service %q created.\n", result.GetObjectMeta().GetName())

	return result.GetObjectMeta().GetName(), err
}

func UpdateIngressCodeServer(clientset *kubernetes.Clientset, namespace string, randomNumber int, ingressName, serviceName string) {
	// Retrieve the ingress
	ingressClient := clientset.NetworkingV1().Ingresses(namespace)
	ingress, err := ingressClient.Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting ingress: %v\n", err)
		return
	}

	pathType := networkingv1.PathTypePrefix

	// New rule created
	newRule := networkingv1.IngressRule{
		Host: "code-server-" + strconv.FormatInt(int64(randomNumber), 10),
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     "/",
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: serviceName,
								Port: networkingv1.ServiceBackendPort{
									Number: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Add new rule in Ingress
	ingress.Spec.Rules = append(ingress.Spec.Rules, newRule)

	// Update Ingress inside the cluster
	updatedIngress, err := ingressClient.Update(context.TODO(), ingress, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error updating ingress: %v\n", err)
		return
	}

	fmt.Printf("Ingress %q updated with new rule for service %q.\n", updatedIngress.Name, serviceName)
}

func StartCodeServer(clientset *kubernetes.Clientset, namespace, labelSelector string) {
	randInstance := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := randInstance.Intn(1000)

	StartDeploymentCodeServer(clientset, namespace, randomNumber)

	serviceName, err := StartServiceCodeServer(clientset, namespace, randomNumber)
	if err != nil {
		fmt.Println("Error in function StartServiceCodeServer")
		return
	}

	UpdateIngressCodeServer(clientset, namespace, randomNumber, "myingress", serviceName)
}

func ListCodeServers(clientset *kubernetes.Clientset, namespace, labelSelector string) {
	fmt.Printf("Listing deployments in namespace %q:\n", namespace)
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Problem while trying to list deployments.")
		return
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}

// TODO :
// - Delete the service
// - Update de the ingress -> delete the rule
func StopCodeServer(clientset *kubernetes.Clientset, deploymentName string) {
	fmt.Println("Deleting deployment...")
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Problem while trying to delete the deployment :", deploymentName)
		return
	}
	fmt.Println("Deleted deployment.")
}
