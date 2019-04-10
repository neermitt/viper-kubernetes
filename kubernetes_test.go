package kubernetes_test

import (
	"testing"
	"time"

	"github.com/neermitt/viper-kubernetes"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubernetesConfigManager_Get(t *testing.T) {
	configMap := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "testConfigMap", Namespace: "default"}, BinaryData: map[string][]byte{"Key1": []byte("Value1"), "Key2": []byte("Value2")}}
	kcm := kubernetes.ConfigMapConfigManager{Client: fake.NewSimpleClientset(configMap)}
	data, err := kcm.Get("testConfigMap/Key1")
	require.NoError(t, err)
	require.Equal(t, []byte("Value1"), data)
}

func TestKubernetesConfigManager_Watch(t *testing.T) {
	configMap := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "testConfigMap", Namespace: "default"}, BinaryData: map[string][]byte{"Key1": []byte("Value1"), "Key2": []byte("Value2")}}
	clientset := fake.NewSimpleClientset(configMap)
	kcm := kubernetes.ConfigMapConfigManager{Client: clientset}
	stopChan := make(chan bool)
	defer close(stopChan)
	resp := kcm.Watch("testConfigMap/Key1", stopChan)
	time.Sleep(time.Second)
	updatedConfigMap := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "testConfigMap", Namespace: "default"}, BinaryData: map[string][]byte{"Key1": []byte("Value2"), "Key2": []byte("Value2")}}
	_, err := clientset.CoreV1().ConfigMaps("default").Update(updatedConfigMap)
	require.NoError(t, err)
	response := <-resp

	require.NoError(t, response.Error)
	require.Equal(t, []byte("Value2"), response.Value)
}
