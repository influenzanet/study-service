# Migration helper tool for survey data models

## Intro

In 2022, it was discovered, the way survey versions were stored caused issues with adding new survey updates. The reason behind this is mongoDB's document size limit (~16MB), where certain surveys reached this quite fast.
To resolve the issue, we changed the survey definition data model, to store survey versions not into one document, but each version into its own document.
To migrate existing files and running systems, we created this simple tool to update the data model from the old format to the newly used data models.

## Usage

Currently the migration tool supports two modes:

1. Apply conversion on a mongoDB collection for a specific `instanceID`  and `studyKey`
2. Convert a JSON file containing the survey definition object

To use these, run the tool with the following command line flags:

### DB conversion

```
-mode=DB -instanceID=myInstance -studyKey=myStudy
```

Configuration to the DB is done through the same environment variables, as for using the normal study-service. When calling the tool with the appropriate command line arguments, make sure, all the DB config values are set (e.g., use a wrapper script to set the variables, and passing the command line arguments).

See `run-example.sh` to run the tool. Alternatively, you can prepare an executable file and run it on the target environment similarly.

### JSON file conversion

```
-mode=JSON -input=/path/to/file.json
```

It will produce a file into the current folder with a name `newSurveyHistory.json`. WARNING: currently the tool will simply override this file every time -> move/rename the file manually, if you want to convert other files.
