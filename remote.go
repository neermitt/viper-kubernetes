package kubernetes

import (
	"bytes"
	"io"
	"log"

	"github.com/spf13/viper"
)

type Response struct {
	Value []byte
	Error error
}

// A ConfigManager retrieves and decrypts configuration from a key/value store.
type ConfigManager interface {
	Get(key string) ([]byte, error)
	Watch(key string, stop chan bool) <-chan *Response
}

type remoteConfigProvider struct {
}

func (kcp remoteConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		log.Printf("Failed to create Config Manager: %v", err)
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func (kcp remoteConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		log.Printf("Failed to create Config Manager: %v", err)
		return nil, err
	}
	resp, err := cm.Get(rp.Path())
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return nil, err
	}

	return bytes.NewReader(resp), nil
}

func (kcp remoteConfigProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	cm, err := getConfigManager(rp)
	if err != nil {
		log.Printf("Failed to create Config Manager: %v", err)
		return nil, nil
	}
	quit := make(chan bool)
	quitwc := make(chan bool)
	viperResponsCh := make(chan *viper.RemoteResponse)
	cryptoResponseCh := cm.Watch(rp.Path(), quit)
	// need this function to convert the Channel response form crypt.Response to viper.Response
	go func(cr <-chan *Response, vr chan<- *viper.RemoteResponse, quitwc chan bool, quit chan<- bool) {
		defer close(quit)
		defer close(quitwc)

		for {
			select {
			case <-quitwc:
				quit <- true
				return
			case resp := <-cr:
				if resp.Error != nil {
					log.Printf("Received error in watch event: %v", resp.Error)
				}
				vr <- &viper.RemoteResponse{
					Error: resp.Error,
					Value: resp.Value,
				}

			}

		}
	}(cryptoResponseCh, viperResponsCh, quitwc, quit)

	return viperResponsCh, quitwc
}

func getConfigManager(rp viper.RemoteProvider) (ConfigManager, error) {
	var cm ConfigManager
	var err error

	switch rp.Provider() {
	case "configmap":
		cm, err = NewConfigMapConfigManager(rp.SecretKeyring())
	default:
		cm, err = NewSecretConfigManager(rp.SecretKeyring())
	}
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func init() {
	viper.RemoteConfig = &remoteConfigProvider{}
	viper.SupportedRemoteProviders = []string{"configmap", "secret"}
}
