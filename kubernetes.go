package kubernetes

import (
	"errors"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var defaultNamespace = "default"

type ConfigMapConfigManager struct {
	Client kubernetes.Interface
}

func (kcm ConfigMapConfigManager) Get(key string) ([]byte, error) {
	ns, name, configKey := parsePath(key)
	configMap, err := kcm.Client.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return extractKeyFromConfigMap(configMap, configKey)
}

func (kcm ConfigMapConfigManager) Watch(key string, stop chan bool) <-chan *Response {
	ns, name, configKey := parsePath(key)
	resp := make(chan *Response, 0)

	go func(configMapName string, key string, stop <-chan bool, resp chan *Response) {
		defer close(resp)

		watch, err := kcm.Client.CoreV1().ConfigMaps(ns).Watch(metav1.ListOptions{})
		if err != nil {
			resp <- &Response{Error: err}
		}

		for {
			select {
			case <-stop:
				watch.Stop()
				return
			case event := <-watch.ResultChan():
				configMap := event.Object.(*v1.ConfigMap)
				if configMap.Name == configMapName {
					if data, err := extractKeyFromConfigMap(configMap, key); err != nil {
						resp <- &Response{Error: err}
					} else {
						resp <- &Response{Value: data}
					}
				}
			}
		}

	}(name, configKey, stop, resp)

	return resp
}

func extractKeyFromConfigMap(configMap *v1.ConfigMap, key string) ([]byte, error) {
	if val, ok := configMap.Data[key]; ok {
		return []byte(val), nil
	}

	if val, ok := configMap.BinaryData[key]; ok {
		return val, nil
	}

	return nil, errors.New("Missing file in config map")
}

type SecretConfigManager struct {
	Client kubernetes.Interface
}

func (scm SecretConfigManager) Get(key string) ([]byte, error) {
	ns, name, configKey := parsePath(key)
	secret, err := scm.Client.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if val, ok := secret.Data[configKey]; ok {
		return []byte(val), nil
	}

	return nil, errors.New("Missing file in secret")
}

func (scm SecretConfigManager) Watch(key string, stop chan bool) <-chan *Response {
	ns, name, configKey := parsePath(key)
	resp := make(chan *Response, 0)

	go func(secretName string, key string, stop <-chan bool, resp chan *Response) {
		defer close(resp)

		watch, err := scm.Client.CoreV1().Secrets(ns).Watch(metav1.ListOptions{})
		if err != nil {
			resp <- &Response{Error: err}
		}

		for {
			select {
			case <-stop:
				watch.Stop()
				return
			case event := <-watch.ResultChan():
				secret := event.Object.(*v1.Secret)
				if secret.Name == secretName {
					if val, ok := secret.Data[configKey]; ok {
						resp <- &Response{Value: val}
					} else {
						resp <- &Response{Error: errors.New("Missing file in secret")}
					}
				}
			}
		}

	}(name, configKey, stop, resp)

	return resp
}

func parsePath(path string) (ns string, name string, key string) {
	values := strings.Split(path, "/")
	switch len(values) {
	case 3:
		ns = values[0]
		name = values[1]
		key = values[2]
	case 2:
		ns = defaultNamespace
		name = values[0]
		key = values[1]
	default:
	}
	return
}

func NewConfigMapConfigManager(configPath string) (ConfigManager, error) {
	config, err := GetConfigFromReader(configPath)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &ConfigMapConfigManager{Client: clientset}, nil
}

func NewSecretConfigManager(configPath string) (ConfigManager, error) {
	config, err := GetConfigFromReader(configPath)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &SecretConfigManager{Client: clientset}, nil
}

func GetConfigFromReader(configPath string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here
	if configPath != "" {
		loadingRules.Precedence = append([]string{configPath}, loadingRules.Precedence...)
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	// if you want to change override values or bind them to flags, there are methods to help you

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	ns, _, err := kubeConfig.Namespace()
	if err == nil {
		defaultNamespace = ns
	}
	return kubeConfig.ClientConfig()
}
