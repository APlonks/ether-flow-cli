package main

import (
	"flag"
	"fmt"
	codeserver "kube-flow/internal/code-server"
	"kube-flow/internal/state"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	var (
		kubeconfigPath  string
		kubeconfig      *string
		choice          int
		config          *rest.Config
		clientset       *kubernetes.Clientset
		deploymentName  string
		namespace       string
		labelProjet     string
		labelCodeServer string
		err             error
	)

	// This values need to be configure under .env and will be replace with values.yaml from the helm
	namespace = "default"
	labelProjet = "project=ethernetes"
	labelCodeServer = "type=code-server"

	kubeconfigPath = "./client.config"

	kubeconfig = flag.String("kubeconfig", kubeconfigPath, "absolute path")
	flag.Parse()

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for {
		fmt.Println("Choose what do u want to do:")
		fmt.Println("0: List all ressources from ethernetes")
		fmt.Println("1: List all deployments from ethernetes")
		fmt.Println("2: List all services from ethernetes")
		fmt.Println("3: List all pods from ethernetes")
		fmt.Println("7: Create code-server")
		fmt.Println("8: List code-server Deployments")
		fmt.Println("9: Remove code-server")

		fmt.Print("\nYour choice: ")
		fmt.Scanf("%d", &choice)
		switch choice {
		case 0:
			fmt.Println("Going to List all ressources from ethernetes")
			fmt.Println()
			state.AllEthernetesRessources(clientset)
			fmt.Println()
		case 1:
			fmt.Println("Going to List all ressources from ethernetes")
			fmt.Println()
			state.ListDeployments(clientset, namespace, labelProjet)
			fmt.Println()
		case 2:
			fmt.Println("Going to List all services from ethernetes")
			fmt.Println()
			state.ListServices(clientset, namespace, labelProjet)
			fmt.Println()
		case 3:
			fmt.Println("Going to List all pods from ethernetes")
			fmt.Println()
			state.ListPods(clientset, namespace, labelProjet)
			fmt.Println()
		case 7:
			fmt.Println("Going to Start Code Server deploying a deployment + cluster ip service.")
			fmt.Println()
			codeserver.StartCodeServer(clientset, namespace, labelProjet)
			fmt.Println()
		case 8:
			fmt.Println("Going to List Code Server")
			fmt.Println()
			codeserver.ListCodeServers(clientset, namespace, labelProjet+","+labelCodeServer)
			fmt.Println()
		case 9:
			fmt.Println("Going to Delete a deployment")
			fmt.Print("Enter the name of the deployment to delete: ")
			fmt.Scanf("%s", &deploymentName)
			fmt.Println("The value of deployment Name: ", deploymentName)
			codeserver.StopCodeServer(clientset, deploymentName)
			fmt.Println()
		default:
			fmt.Println("Choose a valid number")
		}
	}

	// codeserver.StartCodeServer(clientset)

	// // List Deployments
	// prompt()
	// codeserver.ListCodeServers(clientset)

	// Delete Deployment
	// prompt()

}

// func prompt() {
// 	fmt.Printf("-> Press Return key to continue.")
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		break
// 	}
// 	if err := scanner.Err(); err != nil {
// 		panic(err)
// 	}
// 	fmt.Println()
// }
