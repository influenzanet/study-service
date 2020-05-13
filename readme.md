# Study Service
This backend service of the Influenzanet is responsible to manage a study's lifecycle and participants' state through these studies.




## Test
With a running go setup, you can use the command
```
make test
```
to execute the test script. Makefile expects the test script to be at test/test.sh. The test script could contain DB secrets therefore are not added to this git repository. An example [test script](test/example_test_srcipt.sh) can be found in the `test` folder.

Currently the tests also require a working database connection to a mongoDB instance.

## Build
### Docker
Dockerfile(s) are located in `build/docker`. The default Dockerfile is using a multistage build and create a minimal image base on `scratch`.
To trigger the build process using the default docker file call:
```
make docker
```
This will use the current git state (last tag plus if commits since then the commit hash) to tag the docker image.

#### Contribute:
Feel free to create your own Dockerfile (e.g. compiling and deploying to specific target images), eventually others may need the same.
You can create a pull request with adding the Dockerfile into `build/docker` with a good name that it can be identified well, and add a short description to `build/docker/readme.md` about the purpose and speciality of it.

An example to run your created docker image - with the set environment variables - can be found [here](build/docker/example).

## Develop
### API
API definition and data models for the gRPC service can be found in the `api` folder.
To compile the definitions, use:
```
make api
```
This generates the go package into `pgk/api`.
