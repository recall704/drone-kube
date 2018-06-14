package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type (
	// Repo ...
	Repo struct {
		Owner string
		Name  string
	}
	// Build ...
	Build struct {
		Tag     string
		Event   string
		Number  int
		Commit  string
		Ref     string
		Branch  string
		Author  string
		Status  string
		Link    string
		Started int64
		Created int64
	}
	// Job ...
	Job struct {
		Started int64
	}
	// Config ...
	Config struct {
		Ca        string
		Server    string
		Token     string
		Namespace string
		Kind      string
		Workload  string
	}
	// Plugin ...
	Plugin struct {
		Repo   Repo
		Build  Build
		Config Config
		Job    Job
	}
)

// Exec ...
func (p Plugin) Exec() error {

	if p.Config.Server == "" {
		log.Fatal("KUBE_SERVER is not defined")
	}
	if p.Config.Token == "" {
		log.Fatal("KUBE_TOKEN is not defined")
	}
	if p.Config.Ca == "" {
		log.Fatal("KUBE_CA is not defined")
	}
	if p.Config.Namespace == "" {
		p.Config.Namespace = "default"
	}
	if p.Config.Kind == "" {
		log.Fatal("KUBE_KIND, or kind must be defined")
	}
	if p.Config.Workload == "" {
		log.Fatal("KUBE_WORKLOAD, or workload must be defined")
	}

	// create k8s client
	clientset, err := p.createKubeClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = deletePod(clientset, p.Config.Kind, p.Config.Namespace, p.Config.Workload)
	if err != nil {
		log.Fatal(err.Error())
	}
	return err
}

// delete pod by labels
func deletePod(clientset *kubernetes.Clientset, kind, namespace, name string) error {
	var labelSelector labels.Selector
	var err error
	if kind == "Deployment" {
		d, err := clientset.Extensions().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		labelSelector, err = metav1.LabelSelectorAsSelector(d.Spec.Selector)
		if err != nil {
			return err
		}
	} else if kind == "StatefulSet" {
		st, err := clientset.Apps().StatefulSets(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		labelSelector, err = metav1.LabelSelectorAsSelector(st.Spec.Selector)
		if err != nil {
			return err
		}
	} else if kind == "DaemonSet" {
		ds, err := clientset.Extensions().DaemonSets(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		labelSelector, err = metav1.LabelSelectorAsSelector(ds.Spec.Selector)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid kind, kind must be Deployment / StatefulSet / DaemonSet")
	}

	// delete pod by labels, workload will be restart
	err = clientset.CoreV1().Pods(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{
		FieldSelector: fields.Everything().String(),
		LabelSelector: labelSelector.String(),
	})
	return err
}

// create the connection to kubernetes based on parameters passed in.
// the kubernetes/client-go project is really hard to understand.
func (p Plugin) createKubeClient() (*kubernetes.Clientset, error) {

	ca, err := base64.StdEncoding.DecodeString(p.Config.Ca)
	config := clientcmdapi.NewConfig()
	config.Clusters["drone"] = &clientcmdapi.Cluster{
		Server: p.Config.Server,
		CertificateAuthorityData: ca,
	}
	config.AuthInfos["drone"] = &clientcmdapi.AuthInfo{
		Token: p.Config.Token,
	}

	config.Contexts["drone"] = &clientcmdapi.Context{
		Cluster:  "drone",
		AuthInfo: "drone",
	}
	//config.Clusters["drone"].CertificateAuthorityData = ca
	config.CurrentContext = "drone"

	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*config, "drone", &clientcmd.ConfigOverrides{}, nil)
	actualCfg, err := clientBuilder.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	return kubernetes.NewForConfig(actualCfg)
}
