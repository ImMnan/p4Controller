package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// This package function will be responsible for deploying & deleting the k8s resources
// READ/GET, CREATE/DEPLOY, DELETE

type StsData struct {
	StsName  string
	PodType  string
	Type     string
	PodName  string
	PodPort  int
	Services string
	Init     bool
}

func stsRead(cs *ClientSet) (sts []StsData, err error) {
	ns := os.Getenv("WORKING_NAMESPACE")
	if ns == "" {
		ns = "default"
	}
	// List StatefulSets with label selectors

	labelSelector := "app=p4d,managed-by=p4controller"
	stss, err := cs.clientset.AppsV1().StatefulSets(ns).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list StatefulSets: %w", err)
	}

	for _, s := range stss.Items {
		// Get the statefulset name
		stsName := s.Name

		// (server label is not used directly)
		// For each replica, get the Pod name (format: stsName-index)
		replicas := int32(1)
		if s.Spec.Replicas != nil {
			replicas = *s.Spec.Replicas
		}
		// Get the 'server' label as Type (master/replica)
		typ := s.Labels["server"]
		var services string
		var Podport int
		var init bool
		if len(s.Spec.Template.Spec.Containers) > 0 {
			container := s.Spec.Template.Spec.Containers[0]
			for _, env := range container.Env {
				if env.Name == "P4C_PORT" {
					fmt.Sscanf(env.Value, "%d", &Podport)
				}
				if env.Name == "P4C_SERVICES" {
					services = env.Value
				}
				if env.Name == "P4C_INIT" && (env.Value == "true" || env.Value == "1") {
					init = true
				}
			}
		}
		for i := int32(0); i < replicas; i++ {
			podName := fmt.Sprintf("%s-%d", stsName, i)
			sts = append(sts, StsData{
				StsName: stsName,
				PodType: typ,
				//	Type:     exType,
				PodName:  podName,
				PodPort:  Podport,
				Services: services,
				Init:     init,
			})
		}
	}
	return sts, nil
}

func stsDeployer(cs *ClientSet, stsList []StsData) error {
	templatePath := "/etc/p4controller/deployments/replica.yaml"
	file, err := os.Open(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open template %s: %w", templatePath, err)
	}
	defer file.Close()
	templateBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	for _, sts := range stsList {
		// Unmarshal YAML to StatefulSet struct
		var statefulSet appsv1.StatefulSet
		if err := yaml.Unmarshal(templateBytes, &statefulSet); err != nil {
			return fmt.Errorf("failed to unmarshal template YAML: %w", err)
		}

		// Set metadata.name
		statefulSet.ObjectMeta.Name = sts.StsName

		// Set selector.matchLabels.server and template.metadata.labels.server
		if statefulSet.Spec.Selector != nil {
			if statefulSet.Spec.Selector.MatchLabels == nil {
				statefulSet.Spec.Selector.MatchLabels = map[string]string{}
			}
			statefulSet.Spec.Selector.MatchLabels["server"] = sts.PodType
			statefulSet.Spec.Selector.MatchLabels["podName"] = sts.PodName
		}
		if statefulSet.Spec.Template.ObjectMeta.Labels == nil {
			statefulSet.Spec.Template.ObjectMeta.Labels = map[string]string{}
		}
		statefulSet.Spec.Template.ObjectMeta.Labels["server"] = sts.PodType
		statefulSet.Spec.Template.ObjectMeta.Labels["podName"] = sts.PodName

		// Update container env and ports
		if len(statefulSet.Spec.Template.Spec.Containers) > 0 {
			ctr := &statefulSet.Spec.Template.Spec.Containers[0]
			// Remove duplicate envs and set required ones
			var newEnvs []corev1.EnvVar
			for _, env := range ctr.Env {
				if env.Name != "P4C_PORT" && env.Name != "P4C_SERVICES" && env.Name != "P4C_INIT" {
					newEnvs = append(newEnvs, env)
				}
			}
			newEnvs = append(newEnvs,
				corev1.EnvVar{Name: "P4C_PORT", Value: fmt.Sprintf("%d", sts.PodPort)},
				corev1.EnvVar{Name: "P4C_SERVICES", Value: sts.Services},
				corev1.EnvVar{Name: "P4C_INIT", Value: fmt.Sprintf("%v", sts.Init)},
			)
			ctr.Env = newEnvs

			// Update ports
			var newPorts []corev1.ContainerPort
			for _, port := range ctr.Ports {
				if port.Name != "p4d-port" {
					newPorts = append(newPorts, port)
				}
			}
			newPorts = append(newPorts, corev1.ContainerPort{
				Name:          "p4d-port",
				ContainerPort: int32(sts.PodPort),
			})
			ctr.Ports = newPorts
		}

		// Remove duplicate labels in metadata.labels
		if statefulSet.ObjectMeta.Labels == nil {
			statefulSet.ObjectMeta.Labels = map[string]string{}
		}
		statefulSet.ObjectMeta.Labels["server"] = sts.Type
		statefulSet.ObjectMeta.Labels["podName"] = sts.PodName

		// Set namespace if not set
		ns := statefulSet.ObjectMeta.Namespace
		if ns == "" {
			ns = os.Getenv("WORKING_NAMESPACE")
			if ns == "" {
				ns = "default"
			}
			statefulSet.ObjectMeta.Namespace = ns
		}

		// Create the StatefulSet
		_, err := cs.clientset.AppsV1().StatefulSets(ns).Create(
			context.Background(),
			&statefulSet,
			metav1.CreateOptions{},
		)
		if err != nil {
			// If already exists, skip or update as needed (optional)
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return fmt.Errorf("failed to create StatefulSet %s: %w", sts.StsName, err)
		}
	}
	return nil
}
