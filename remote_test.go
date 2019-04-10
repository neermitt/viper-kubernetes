package kubernetes_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestViperRemoteConfig_FromConfigMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Ignore kubernetes tests")
	}

	v := viper.New()
	err := v.AddRemoteProvider("configmap", "dummy", "test/conf.yaml")
	require.NoError(t, err)
	v.SetConfigFile("config.yaml")
	err = v.ReadRemoteConfig()
	require.NoError(t, err)

	name := v.GetString("app.name")

	require.Equal(t, "test", name)

}

func TestViperRemoteConfig_FromSecret(t *testing.T) {
	if testing.Short() {
		t.Skip("Ignore kubernetes tests")
	}
	v := viper.New()
	err := v.AddRemoteProvider("secret", "dummy", "test/config.yaml")
	require.NoError(t, err)
	v.SetConfigFile("config.yaml")
	err = v.ReadRemoteConfig()
	require.NoError(t, err)

	name := v.GetString("app.name")

	require.Equal(t, "test2", name)

}
