package main

import (
	"flag"
	"fmt"
	codeserver "kube-flow/internal/code-server"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	var (
		kubeconfigPath string
		kubeconfig     *string
		choice         int
		config         *rest.Config
		clientset      *kubernetes.Clientset
		deploymentName string
		err            error
	)

	kubeconfigPath = "./client.config"
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
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
		fmt.Println("1: Create code-server")
		fmt.Println("2: Liste Deployment")
		fmt.Println("3: Remove code-server")

		fmt.Scanf("%d", &choice)
		switch choice {
		case 1:
			fmt.Println("Going to Start Code Server")
			codeserver.StartCodeServer(clientset)
			fmt.Println()
		case 2:
			fmt.Println("Going to List Code Server")
			codeserver.ListCodeServers(clientset)
			fmt.Println()
		case 3:
			fmt.Println("Going to Delete Code Server")
			fmt.Print("Enter the name of the deployment to delete: ")
			fmt.Scanf("%s", &deploymentName)
			fmt.Println("The value of deployment Name: ", deploymentName)
			codeserver.StopCodeServer(clientset, deploymentName)
			fmt.Println()
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
