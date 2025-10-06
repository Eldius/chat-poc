package config

import (
	"time"

	"github.com/spf13/viper"
)

func GetDBUser() string {
	return viper.GetString(DBUserPropKey)
}

func GetDBPass() string {
	return viper.GetString(DBPassPropKey)
}

func GetDBHost() string {
	return viper.GetString(DBHostPropKey)
}

func GetDBPort() string {
	return viper.GetString(DBPortPropKey)
}

func GetDBName() string {
	return viper.GetString(DBNamePropKey)
}

func GetBedrockInferenceModel() string {
	return viper.GetString(BedrockInferenceModelProp.Key)
}

func GetBedrockEmbeddingModel() string {
	return viper.GetString(BedrockEmbeddingModelProp.Key)
}

func GetBedrockRegion() string {
	return viper.GetString(BedrockRegionProp.Key)
}

func GetBedrockInferenceTemperature() float64 {
	return viper.GetFloat64(BedrockInferenceTemperatureProp.Key)
}

func GetBedrockInferenceMaxIterations() int {
	return viper.GetInt(BedrockInferenceMaxIterationsProp.Key)
}

func GetBedrockInferenceTopK() int {
	return viper.GetInt(BedrockInferenceTopKProp.Key)
}

func GetBedrockInferenceTopP() float64 {
	return viper.GetFloat64(BedrockInferenceTopPProp.Key)
}

func GetBedrockCacheEnabled() bool {
	return viper.GetBool(BedrockCacheEnabledProp.Key)
}

func GetBedrockCachePersistTimeout() time.Duration {
	return viper.GetDuration(BedrockCachePersistTimeoutProp.Key)
}

func GetCacheDBPath() string {
	return viper.GetString(CacheDBPathProp.Key)
}

func GetAPIPort() string {
	return viper.GetString(APIPortProp.Key)
}
