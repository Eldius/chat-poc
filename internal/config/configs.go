package config

import "github.com/spf13/viper"

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
