package config

const (
	ENV_GRPC_MAX_MSG_SIZE           = "GRPC_MAX_MSG_SIZE"
	ENV_PERSISTENCE_STORE_ROOT_PATH = "PERSISTENCE_STORE_ROOT_PATH"
)

const (
	defaultGRPCMaxMsgSize           = 4194304
	defaultPersistenceStoreRootPath = "files"
)
