export STUDY_DB_CONNECTION_STR="<connection string without prefix and auth data>"
export STUDY_DB_CONNECTION_PREFIX="<connection string prefix (e.g. +srv for atlas)>"
export STUDY_DB_USERNAME="<db username>"
export STUDY_DB_PASSWORD="<db password>"

export GLOBAL_DB_CONNECTION_STR="<connection string without prefix and auth data>"
export GLOBAL_DB_CONNECTION_PREFIX="<connection string prefix (e.g. +srv for atlas)>"
export GLOBAL_DB_USERNAME="<db username>"
export GLOBAL_DB_PASSWORD="<db password>"

export DB_TIMEOUT=30 # seconds until connection times out
export DB_IDLE_CONN_TIMEOUT=45 # terminate idle connection after seconds
export DB_MAX_POOL_SIZE=8
export DB_DB_NAME_PREFIX="<db name prefix used in the test>" # DB names will be then > <DB_PREFIX>+"hard-coded-db-name-as-we-need-it"

export STUDY_SERVICE_LISTEN_PORT=5203
export STUDY_GLOBAL_SECRET="<global secret key for participant id encryption>"
export STUDY_TIMER_EVENT_FREQUENCY=30
export STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN=5
export STUDY_TIMER_EVENT_CHECK_INTERVAL_VAR=3

go test  ./... $1
