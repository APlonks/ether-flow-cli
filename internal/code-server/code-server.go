package codeserver

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"kube-flow/internal/utils"
)

// TODO :
// - Update the the ingress to expose the codeserver
// - Also configure random number for labels completion
// - Return the medata label of the pod to give it to the service
// It is necessary to deploy more than one deployment
func StartDeploymentCodeServer(clientset *kubernetes.Clientset) {
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "code-server-deployment",
			Labels: map[string]string{
				"app":     "code-server-deployment",
				"project": "ethernetes",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "code-server-pod",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "code-server-pod",
						"project": "ethernetes",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "code-server-container",
							Image: "docker.io/codercom/code-server:4.22.1-ubuntu",
							Env: []apiv1.EnvVar{
								{
									Name:  "DOCKER_USER",
									Value: "user",
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http-code",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("256Mi"),
									apiv1.ResourceCPU:    resource.MustParse("500m"),
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config-code-server",
									MountPath: "/home/coder/.config",
								},
								{
									Name:      "project-storage",
									MountPath: "/home/coder/project",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "config-code-server",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "project-storage",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("The deployment didn't work.")
		return
	}
	fmt.Printf("Deployment created %q.\n", result.GetObjectMeta().GetName())
}

func StartServiceCodeServer(clientset *kubernetes.Clientset) {
	serviceName := "code-server-service"

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"app":     "code-server",
				"project": "ethernetes",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": "code-server-pod", // Label to select the pod
			},
			Ports: []apiv1.ServicePort{
				{
					Protocol:   apiv1.ProtocolTCP,
					Port:       80,                   // Service's port
					TargetPort: intstr.FromInt(8080), // Pod's port
				},
			},
		},
	}

	fmt.Println("Creating service for code-server...")
	result, err := clientset.CoreV1().Services(apiv1.NamespaceDefault).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating service: %v\n", err)
		return
	}
	fmt.Printf("Service %q created.\n", result.GetObjectMeta().GetName())
}

func UpdateIngressCodeServer(clientset *kubernetes.Clientset, ingressName, namespace, serviceName, hostName string) {
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
		Host: hostName,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     "/code-server",
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
	StartDeploymentCodeServer(clientset)
	StartServiceCodeServer(clientset)
	UpdateIngressCodeServer(clientset, "myingress", namespace, "code-server-service", "code-server")
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
// - Update de the ingress
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
