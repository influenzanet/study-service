# Study Service
This backend service of the Influenzanet is responsible to manage a study's lifecycle.

## Test
To perform tests, some environmental variables has to be set. After installing the dependencies, you can add a script to initialize these variables and perform tests. This script is not included in this repository, since it contains secret infos, like DB password.
Currently the tests also require a working database connection to a mongoDB instance.

Here is an example how such a script could look like:

```sh
export DB_CONNECTION_STR="<mongo db connection string without prefix and auth infos>"
export DB_USERNAME="<username for mongodb auth>"
export DB_PASSWORD="<password for mongodb auth>"
export DB_PREFIX="+srv" # e.g. "" (emtpy) or "+srv"
export DB_TIMEOUT=30 # seconds until connection times out
export DB_IDLE_CONN_TIMEOUT=45 # terminate idle connection after seconds
export DB_MAX_POOL_SIZE=8
export DB_DB_NAME_PREFIX="<DB_PREFIX>" # DB names will be then > <DB_PREFIX>+"hard-coded-db-name-as-we-need-it"
export STUDY_SERVICE_LISTEN_PORT=5203
export STUDY_GLOBAL_SECRET="<global secret key - part 1 to encrypt userID>"

go test  ./...
```