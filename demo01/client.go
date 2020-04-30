package demo01
import (
	"flag"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

var (
	restConfig   *rest.Config
	clientset    *kubernetes.Clientset
	err 		 error
	kubeconfig 	 *string

)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func InitClient() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return clientset, nil
}


func GetRestConf() (restConf *rest.Config, err error) {
	var (
		kubeconfig []byte
	)
	FilePath := filepath.Join(homeDir(), ".kube", "config")
	// 读kubeconfig文件
	if kubeconfig, err = ioutil.ReadFile(FilePath); err != nil {
		goto END
	}
	// 生成rest client配置
	if restConf, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig); err != nil {
		goto END
	}
END:
	return
}
