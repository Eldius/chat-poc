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
	BedrockCacheDBPathProp            setup.Prop = setup.Prop{Key: "cache.path", Value: ".db/cache.db"}
	BedrockChatMemoryDBPathProp       setup.Prop = setup.Prop{Key: "bedrock.chat.memory.db_file", Value: ".db/chat.db"}

	ConfluenceAuthRedirectURLProp     setup.Prop = setup.Prop{Key: "confluence.auth.redirect_url", Value: "http://localhost:9999/auth/result"}
	ConfluenceAuthURLProp             setup.Prop = setup.Prop{Key: "confluence.auth.url", Value: "https://auth.atlassian.com/authorize"}
	ConfluenceAuthResponseTypeProp    setup.Prop = setup.Prop{Key: "confluence.auth.response_type", Value: "code"}
	ConfluenceAuthAudienceProp        setup.Prop = setup.Prop{Key: "confluence.auth.audience", Value: "api.atlassian.com"}
	ConfluenceAuthPromptProp          setup.Prop = setup.Prop{Key: "confluence.auth.prompt", Value: "consent"}
	ConfluenceAuthRefreshTokenURLProp setup.Prop = setup.Prop{Key: "confluence.auth.refresh_tokens", Value: "https://auth.atlassian.com/oauth/token"}
	ConfluenceBaseURLProp             setup.Prop = setup.Prop{Key: "confluence.base_url", Value: "https://confluence.atlassian.com"}
	ConfluenceClientIDProp            setup.Prop = setup.Prop{Key: "confluence.client_id", Value: "1234567890"}
	ConfluenceClientSecretProp        setup.Prop = setup.Prop{Key: "confluence.client_secret", Value: "secret"}
	ConfluenceScopesProp              setup.Prop = setup.Prop{Key: "confluence.scopes", Value: []string{
		"read:confluence-space.summary",
		"read:confluence-props",
		"read:confluence-content.all",
		"read:confluence-content.summary",
		"search:confluence",
		"read:confluence-user",
		"read:confluence-groups",
		"readonly:content.attachment:confluence",
	}}

	APIPortProp = setup.Prop{Key: "api.port", Value: "8080"}
)
