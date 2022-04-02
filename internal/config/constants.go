package config

const (
	ENV_GRPC_MAX_MSG_SIZE             = "GRPC_MAX_MSG_SIZE"
	ENV_PERSISTENCE_STORE_ROOT_PATH   = "PERSISTENCE_STORE_ROOT_PATH"
	ENV_PERSISTENCE_MAX_FILE_SIZE     = "PERSISTENCE_STORE_MAX_FILE_SIZE"
	ENV_EXTERNAL_SERVICES_CONFIG_PATH = "EXTERNAL_SERVICES_CONFIG_PATH"
)

const (
	defaultGRPCMaxMsgSize           = 4194304
	defaultPersistenceStoreRootPath = "files"
	maxParticipantFileSize          = 1 << 25
)
