package kor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func ProcessNamespaceDeployments(clientset kubernetes.Interface, namespace string) ([]UnusedResource, error) {
	deploymentsList, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var deploymentsWithoutReplicas []UnusedResource

	for _, deployment := range deploymentsList.Items {

		if deployment.Labels["kor/used"] == "true" {
			continue
		}

		if deployment.Labels["kor/used"] == "false" {
			unusedDeployment := UnusedResource{
				Name:   deployment.Name,
				Reason: "Label kor/used set to false",
			}
			deploymentsWithoutReplicas = append(deploymentsWithoutReplicas, unusedDeployment)
			continue
		}

		if *deployment.Spec.Replicas == 0 {
			unusedDeployment := UnusedResource{
				Name:   deployment.Name,
				Reason: "No resource using this Deployment",
			}
			deploymentsWithoutReplicas = append(deploymentsWithoutReplicas, unusedDeployment)
		}
	}

	return deploymentsWithoutReplicas, nil
}

func GetUnusedDeployments(includeExcludeLists IncludeExcludeLists, clientset kubernetes.Interface, slackOpts SlackOpts) {
	namespaces := SetNamespaceList(includeExcludeLists, clientset)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutputNew(namespace, diff, "Deployments")

		outputBuffer.WriteString(output)
		outputBuffer.WriteString("\n")
	}

	if slackOpts != (SlackOpts{}) {
		if err := SendToSlack(SlackMessage{}, slackOpts, outputBuffer.String()); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send message to slack: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(outputBuffer.String())
	}
}

func GetUnusedDeploymentsStructured(includeExcludeLists IncludeExcludeLists, clientset kubernetes.Interface, outputFormat string) (string, error) {
	namespaces := SetNamespaceList(includeExcludeLists, clientset)
	response := make(map[string]map[string][]UnusedResource)

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		resourceMap := make(map[string][]UnusedResource)
		resourceMap["Deployments"] = diff
		response[namespace] = resourceMap
	}

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}

	if outputFormat == "yaml" {
		yamlResponse, err := yaml.JSONToYAML(jsonResponse)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		return string(yamlResponse), nil
	} else {
		return string(jsonResponse), nil
	}
}
