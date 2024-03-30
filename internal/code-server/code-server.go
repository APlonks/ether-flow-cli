package codeserver

import (
	"bytes"
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
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
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

func RetrieveDataCodeServer(clientset *kubernetes.Clientset, namespace string, randomNumber int, config *rest.Config) (string, error) {

	fmt.Println("1")

	// Construire un sélecteur de labels pour trouver le Pod spécifique
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"number": fmt.Sprintf("%d", randomNumber)}}
	listOptions := metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()}

	// Trouver le Pod en utilisant le sélecteur de labels
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		return "", err
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found with the 'number' label set to %d", randomNumber)
	}

	fmt.Println("2")

	// Utiliser le premier Pod pour l'exemple
	podName := pods.Items[0].Name

	// La commande à exécuter dans le conteneur.
	cmd := []string{"cat", "/home/coder/.config/code-server/config.yaml"}
	// containerName := pods.Items[0].Spec.Containers[0].Name // Supposer le premier conteneur

	// Configuration pour l'exécution de la commande
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&apiv1.PodExecOptions{
			Command: cmd,
			// Container: containerName,
			Stdin:  false,
			Stdout: true,
			Stderr: true,
			// TTY:    true,
		}, scheme.ParameterCodec)

	fmt.Println("3")

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		fmt.Printf("Error while creating SPDYExecutor: %v\n", err)
		return "", err
	}

	fmt.Println("4")

	// Buffers pour capturer les sorties standard et d'erreur
	var stdout, stderr bytes.Buffer
	// Calling Sleep method
	time.Sleep(1 * time.Second)
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		fmt.Printf("Error in exec.Stream: %v\n", err)
		return "", err
	}

	fmt.Println("5")

	fmt.Printf("Command output:\nSTDOUT: %s\nSTDERR: %s\n", stdout.String(), stderr.String())

	// Retourner la sortie standard comme résultat
	return stdout.String(), nil
}

func StartCodeServer(clientset *kubernetes.Clientset, namespace, labelSelector string, config *rest.Config) {
	randInstance := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := randInstance.Intn(1000)

	StartDeploymentCodeServer(clientset, namespace, randomNumber)

	serviceName, err := StartServiceCodeServer(clientset, namespace, randomNumber)
	if err != nil {
		fmt.Println("Error in function StartServiceCodeServer")
		return
	}

	UpdateIngressCodeServer(clientset, namespace, randomNumber, "myingress", serviceName)

	// Calling Sleep method
	time.Sleep(5 * time.Second)

	// Printed after sleep is over
	fmt.Println("Sleep Over.....")

	RetrieveDataCodeServer(clientset, namespace, randomNumber, config)
	fmt.Println("Done")

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
func StopCodeServer(clientset *kubernetes.Clientset, namespace string, codeServerNumber int) {

	deleteDeploymentCodeServer(clientset, namespace, "code-server-deployment-"+strconv.FormatInt(int64(codeServerNumber), 10))

	deleteServiceCodeServer(clientset, namespace, "code-server-service-"+strconv.FormatInt(int64(codeServerNumber), 10))

	deleteIngressRule(clientset, namespace, "myingress", "code-server-"+strconv.FormatInt(int64(codeServerNumber), 10))
}

func deleteDeploymentCodeServer(clientset *kubernetes.Clientset, namespace string, deploymentName string) {
	fmt.Println("Deleting deployment...")
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Problem while trying to delete the code server :", deploymentName)
		return
	}
	fmt.Println("Deleted deployment.")
}

func deleteServiceCodeServer(clientset *kubernetes.Clientset, namespace string, serviceName string) {
	fmt.Printf("Deleting service %s...\n", serviceName)
	serviceClient := clientset.CoreV1().Services(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	if err := serviceClient.Delete(context.TODO(), serviceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Printf("Error while trying to delete the service %s: %v\n", serviceName, err)
		return
	}

	fmt.Printf("Service %s deleted successfully.\n", serviceName)
}

func deleteIngressRule(clientset *kubernetes.Clientset, namespace, ingressName, hostToFind string) error {
	// Récupérer l'objet Ingress
	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}

	// Parcourir les règles pour trouver et supprimer la règle correspondant à l'hôte
	newRules := []v1.IngressRule{}
	for _, rule := range ingress.Spec.Rules {
		if rule.Host != hostToFind {
			newRules = append(newRules, rule) // Conserver les règles qui ne correspondent pas
		}
	}

	// Si la longueur des règles est inchangée, l'hôte n'a pas été trouvé
	if len(newRules) == len(ingress.Spec.Rules) {
		return fmt.Errorf("host %s not found in ingress rules", hostToFind)
	}

	// Mettre à jour les règles dans l'objet Ingress
	ingress.Spec.Rules = newRules

	// Mettre à jour l'Ingress dans le cluster
	_, err = clientset.NetworkingV1().Ingresses(namespace).Update(context.TODO(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress after deleting rule: %v", err)
	}

	fmt.Printf("Ingress rule for host %s deleted successfully from %s.\n", hostToFind, ingressName)
	return nil
}
