


export STUDY_DB_CONNECTION_STR="<connection string>"
export STUDY_DB_USERNAME="<username>"
export STUDY_DB_PASSWORD="<password>"
export STUDY_DB_CONNECTION_PREFIX="<connection string prefix / e.g. +srv>"

export DB_TIMEOUT=30
export DB_IDLE_CONN_TIMEOUT=45
export DB_MAX_POOL_SIZE=8
export DB_DB_NAME_PREFIX="INF_"

# Call go run OR build and call executable instead
go run *.go "$@"