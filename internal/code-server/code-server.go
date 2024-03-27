package codeserver

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	"kube-flow/internal/utils"
)

func StartCodeServer(clientset *kubernetes.Clientset) {
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "code-server-deployment",
			Labels: map[string]string{
				"app": "code-server-deployment",
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
						"app": "code-server-pod",
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
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func ListCodeServers(clientset *kubernetes.Clientset) {
	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Problem while trying to list deployments.")
		return
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}

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
