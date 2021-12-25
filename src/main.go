package main

import (
	"context"
	"encoding/json"
	mux "github.com/gorilla/mux"
	"github.com/spotahome/kooper/v2/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net/http"
	"os"
	"path/filepath"
)

var logger log.Logger
var k8sclient *kubernetes.Clientset

func init() {
	logger = log.NewStd(false)
	k8sclient = getK8sClient()
	tittle()
}

type NsData struct {
	Namespace  string   `json:"namespace"`
	Pods       []string `json:"pods"`
	Ingressess []string `json:"ingresess"`
}

func tittle() {
	logger.Infof("-----------------------------------------------------------------\n")
	logger.Infof("oo          .8888b\n")
	logger.Infof("88   \n")
	logger.Infof("dP 88d888b. 88aaa  .d8888b. 88d888b. 88d8b.d8b. .d8888b. 88d888b.\n")
	logger.Infof("88 88'  `88 88     88'  `88 88'  `88 88'`88'`88 88ooood8 88'  `88\n")
	logger.Infof("88 88    88 88     88.  .88 88       88  88  88 88.  ... 88\n")
	logger.Infof("dP dP    dP dP     `88888P' dP       dP  dP  dP `88888P' dP\n")
	logger.Infof("-----------------------------------------------------------------\n")
	return
}

func getK8sClient() *kubernetes.Clientset {

	// Get k8s client.
	k8scfg, err := rest.InClusterConfig()
	if err != nil {
		// No in cluster? letr's try locally
		logger.Infof("Not in a K8s cluster; let's try locally")
		// https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go
		kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
		k8scfg, err = clientcmd.BuildConfigFromFlags("", kubehome)
		if err != nil {
			logger.Errorf("Error loading kubernetes configuration: %s", err)
			os.Exit(1)
		}
	}

	k8scli, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		logger.Errorf("Error creating kubernetes client: %s\n", err)
		os.Exit(1)
	}

	return k8scli

}

func namespaces(w http.ResponseWriter, _ *http.Request) {

	iNamespaces := k8sclient.CoreV1().Namespaces()
	namespaceList, err := iNamespaces.List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		logger.Errorf("Error getting namespace list: %s\n", err)
		return
	}

	outputSlice := make([]string, len(namespaceList.Items))
	for i, namespace := range namespaceList.Items {
		outputSlice[i] = namespace.Name
	}
	output, err := json.Marshal(outputSlice)
	if err != nil {
		logger.Errorf("Error marshalling namespace list: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func namespaceDataByName(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	outputData := NsData{namespace, podList(namespace), ingressList(namespace)}
	byteOutput, err := json.Marshal(outputData)
	if err != nil {
		logger.Errorf("Error marshalling namespace data by name: %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(byteOutput)

}

func podList(namespace string) []string {

	iPods := k8sclient.CoreV1().Pods(namespace)
	podList, err := iPods.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Errorf("Error marshalling pod list by NS: %s\n", err)
		return make([]string, 0)
	}
	outputSlice := make([]string, len(podList.Items))
	for i, pod := range podList.Items {
		outputSlice[i] = pod.Name
	}
	return outputSlice

}

func podsByNamespace(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	output, err := json.Marshal(podList(namespace))
	if err != nil {
		logger.Errorf("Error marshalling pod list: %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(output)

}

func ingressList(namespace string) []string {

	ingressC := k8sclient.NetworkingV1().Ingresses(namespace)
	ingressList, err := ingressC.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Errorf("Error marshalling namespace list: %s\n", err)
		return make([]string, 0)
	}
	outputSlice := make([]string, len(ingressList.Items))
	for i, ingress := range ingressList.Items {
		outputSlice[i] = ingress.Name
	}
	return outputSlice

}

func ingressByNamespace(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	output, err := json.Marshal(ingressList(namespace))
	if err != nil {
		logger.Errorf("Error marshalling ingresses by NS list: %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(output)

}

func middleware(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		handler.ServeHTTP(w, r)
	})

}

func main() {

	r := mux.NewRouter()
	r.Use(middleware)
	r.HandleFunc("/namespaces", namespaces).Methods("GET")
	r.HandleFunc("/namespaces/{namespace}", namespaceDataByName).Methods("GET")
	r.HandleFunc("/pods/{namespace}", podsByNamespace).Methods("GET")
	r.HandleFunc("/ingresses/{namespace}", ingressByNamespace).Methods("GET")
	r.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("1"))
	}).Methods("GET")
	http.ListenAndServe(":8080", r)

}
