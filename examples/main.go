package main

import (
	"log"
	"net/http"

	_ "github.com/neermitt/viper-kubernetes"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()

	err := v.AddRemoteProvider("secret", "dummy", "test/config.yaml")
	if err != nil {
		log.Printf("Failed to add remote provider %v", err)
	}
	v.SetConfigType("yaml")
	err = v.ReadRemoteConfig()
	if err != nil {
		log.Printf("Failed to read remote config %v", err)
	}

	err = v.WatchRemoteConfigOnChannel()
	if err != nil {
		log.Printf("Failed to watch remote config change %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		s := v.GetString("app.name")
		log.Printf("Name: %s", s)
		_, err := writer.Write([]byte(s))
		if err != nil {
			log.Printf("Failed to write response %v", err)
		}
	})
	log.Fatal(http.ListenAndServe(":80", mux))
}
