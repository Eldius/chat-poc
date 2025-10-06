package config

import "github.com/eldius/initial-config-go/setup"

const (
	AppName = "chat-poc"

	DBPassPropKey = "db.pass"
	DBUserPropKey = "db.user"
	DBHostPropKey = "db.host"
	DBPortPropKey = "db.port"
	DBNamePropKey = "db.name"
)

var (
	Version string

	BedrockRegionProp                 setup.Prop = setup.Prop{Key: "bedrock.region", Value: "us-east-1"}
	BedrockInferenceModelProp         setup.Prop = setup.Prop{Key: "bedrock.inference.model", Value: "anthropic.claude-3-haiku-20240307-v1:0"}
	BedrockEmbeddingModelProp         setup.Prop = setup.Prop{Key: "bedrock.embedding.model", Value: "amazon.titan-embed-text-v1"}
	BedrockInferenceTemperatureProp   setup.Prop = setup.Prop{Key: "bedrock.inference.temperature", Value: "0.8"}
	BedrockInferenceMaxIterationsProp setup.Prop = setup.Prop{Key: "bedrock.inference.chain.max_iterations", Value: 5}
	BedrockInferenceTopKProp          setup.Prop = setup.Prop{Key: "bedrock.inference.top_k", Value: 5}
	BedrockInferenceTopPProp          setup.Prop = setup.Prop{Key: "bedrock.inference.top_p", Value: 0.95}
	BedrockCacheEnabledProp           setup.Prop = setup.Prop{Key: "bedrock.inference.cache.enabled", Value: false}
	BedrockCachePersistTimeoutProp    setup.Prop = setup.Prop{Key: "bedrock.inference.cache.timeout", Value: "15s"}

	CacheDBPathProp setup.Prop = setup.Prop{Key: "cache.path", Value: ".db/cache.db"}

	APIPortProp = setup.Prop{Key: "api.port", Value: "8080"}
)
