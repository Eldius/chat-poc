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

	BedrockRegionProp         setup.Prop = setup.Prop{Key: "bedrock.region", Value: "us-east-1"}
	BedrockInferenceModelProp setup.Prop = setup.Prop{Key: "bedrock.inference.model", Value: "anthropic.claude-3-haiku-20240307-v1:0"}
	BedrockEmbeddingModelProp setup.Prop = setup.Prop{Key: "bedrock.embedding.model", Value: "amazon.titan-embed-text-v1"}
)
