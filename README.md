# viper-kubernetes
Provides Viper Remote Config from Kubernetes ConfigMaps and Secrets

# Usage

```go
package main

import (
	"log"
	"github.com/spf13/viper"
	_ "github.com/neermitt/viper-kubernetes"
)

func main() {
		v := viper.New()
    
    	err := v.AddRemoteProvider("secret", "dummy", "secretname/key")
    	if err != nil {
    		log.Printf("Failed to add remote provider %v", err)
    	}
    	v.SetConfigType("yaml")
    	err = v.ReadRemoteConfig()
    	if err != nil {
    		log.Printf("Failed to read remote config %v", err)
    	}
    	
    	log.Print(v.GetString("my.property.name"))
}
```