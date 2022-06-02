# Changelog

## [v1.1.0] - 2022-06-02

### Added

- Study-engine can be now extended with external logic via configurable calls to external HTTP endpoints. There are two new expressions for this:
  - `EXTERNAL_EVENT_HANDLER`: is a study action that can be used to trigger some externally defined logic. The https response might contain the updated participant state (`pState`) and/or the map of reports to be created (`reportsToCreate`) after the rules have run. Both of these are optional. If not provided the previous participant state is kept.
  - `externalEventEval`: is a study expression that can be used to process the event (e.g. survey responses) externally and retrieve a value, that can be used in the study engine (e.g. determine which survey should be assigned). For `externalEventEval` the return value of the expression (received through the HTTP response) can be interpreted as string (by default) or a float64 (if return type is defined as "float").
Both expression will attempt to send an HTTP POST message with a payload containing the `apiKey`, `participantState`, `eventType`, `studyKey`, `instanceID` and if relevant `surveyResponses`.
  - The expressions requires the following arguments:
    - `serviceName`: this name will be used to look up the URL and API key for the service.
  - To configure an external service, the study-service requires a yaml file containing the list of service configs. The path to such a yaml file can be defined through the environment variable `EXTERNAL_SERVICES_CONFIG_PATH`.
  - Content of the config file is stuctured as:

```yaml
services:
- name: nameOfTheEventHandlerService
  url: https://<url-of-the-event-handler-endpoint>
  apiKey: <API key to authenticate the study service>
- name: secondService
  ...
```

- New study engine action `NOTIFY_RESEARCHER` with arguments for `messageType` (which message template will be used / message subscription topic) and an optional list of key and value pairs (that can be used to populate the payload for the message).
  - Added gRPC endpoints to fetch or update notification subscriptions of a study.

## [v1.0.3] - 2022-04-15

### Changed

- Fix issue in report ignoring logic

## [v1.0.2] - 2022-03-15

### Changed

- Participant state export model includes scheduled participant messages as well.

## [v1.0.1] - 2022-03-10

### Changed

- when updating report data with dType "int", value should be printed as int.

## [v1.0.0] - 2022-03-08

### Added

- Study data model now includes the attribute `idMappingMethod`, which allows to select the method used to convert profile ID into participant ID. This configuration is per study. Currently available methods are: 'aesctr' (default for backwards compatibility), 'sha224', 'sha256', 'same'.
- Include improved logger, using configurable log levels. Use the environment variable `LOG_LEVEL` to select which level should be applied. Possible values are: `debug info warning error`.
- New gRPC endpoints:
  - `GetResponsesFlatJSON`: data exporter to export repsonses in a flat JSON list
  - `RegisterTemporaryParticipant`: create a participant that has no account yet
  - `ConvertTemporaryToParticipant`: convert a temporary participant (or merge) into an active participant
  - `GetAssignedSurveysForTemporaryParticipant`: get assigned surveys for a temporary participant
  - `GetReportsForUser`: retrieve the report history for a user
  - `StreamParticipantStates`: stream participant state list according to the query (admin, researcher)
  - `StreamReportHistory`: stream report list according to the query (admin, researcher)
  - `RemoveConfidentialResponsesForProfiles`: remove data from confidential responses collection for the given profiles of the user
  - `UploadParticipantFile`: upload one participant file
  - `DeleteParticipantFiles`: delete a list of participant files by id
  - `GetParticipantFile`: download one participant file
  - `StreamParticipantFileInfos`: stream file infos for the query
  - `GetParticipantMessages`: get messages for a participant
  - `DeleteMessagesFromParticipant`: delete list of messages for a participant
  - `GetResearcherMessages`: get all researcher messages form all studies
  - `DeleteResearcherMessages`: remove researcher messages with study key and id.
  - `RunRulesForSingleParticipant`: run custom rules to a single participant.
  - `CreateReport`: endpoint to create a participant report, e.g., to use in migration process.

- Participant File Upload endpoint: method to upload files for study participants
- Can configure root path of the persistent storage with PERSISTENCE_STORE_ROOT_PATH env variable.
- Can configure max file size for participant file uploads with PERSISTENCE_STORE_MAX_FILE_SIZE env variable.

- Data exporter:
  - Data exporter can parse composite question titles or option labels (when text contains multiple parts).
  - Data exporter logic to handle question type: "responsive single choice array" and "responsive bipolar likert scale array"
  - Exports now contain two new fixed columns: `ID` (identifying a particular survey submission) and `opened` (containing the POSIX time timestamp at which the client opened the survey).
  - Roles can now be extended with custom names using the scheme `role:customName`. If the role is ommitted (format: `:customName`), the item is still exported, but as an 'unknown' question.
  - `cloze` questions and single/multi-choice options are now exported
  - Can handle `timeInput` question type as a number input (it is basically a number = seconds since 00:00).
  - Can handle `consent` question type.
- New messaging concept:
  - Participant state can store list of messages scheduled for the participant.
  - New study-engine expressions `hasMessageTypeAssigned`, `getMessageNextTime`.
  - New study-engine actions: `ADD_MESSAGE`, `REMOVE_MESSAGES_BY_TYPE`, `REMOVE_ALL_MESSAGES`.
  - Added DB method to remove messages by id from participant state
  - New gRPC endpoint to fetch messages for a participant and trigger deletion of messages from a participant's message list.
- Study engine:
  - `hasParticipantFlagKey` new expression, that will check if a participant has a flag attribute with a specific key, where the value is not checked.
  - `getParticipantFlagValue` new expression, that will retrieve the value of a participant flag with a given key.
  - `hasResponseKey` new expression, that will return true if the question has a response that contains a key at the specified path (e.g.: T0.Q1, rg.scg.1.b)
  - `getSelectedKeys` to retrieve the list of keys from a response slot, encoded as a string, separated by semicolons `;`
  - `hasResponseKeyWithValue` new expression, that will return true if the question has a response that contains a key at the specified path with the given value (e.g.: T0.Q1, rg.scg.1.b, value)
  - New actions for working with reports: `INIT_REPORT, UPDATE_REPORT_DATA, REMOVE_REPORT_DATA, CANCEL_REPORT`. At the start o the study event, a map of reports is initalised (empty), and during the event, actions can create one or more report entries. Reports that are in this map at the end of the event will be saved to the database.
  - implemented new expressions to handle merge event (when two participant states should be merged - convert temporary participant when participant already exists)
  - New actions to remove confidential responses: `REMOVE_CONFIDENTIAL_RESPONSE_BY_KEY`, `REMOVE_ALL_CONFIDENTIAL_RESPONSES`
  - New action: `START_NEW_STUDY_SESSION`. Survey responses will include the session attribute, so the study can simply group batch of responses for a participant.
- New prefill method `PREFILL_SLOT_WITH_VALUE`

### Changed

- The `metaInit`, `metaDisplayed`, and `metaResponse` columns are now exported as JSON arrays of int64 POSIX time timestamps.
- `GET_LAST_SURVEY_ITEM` survey prefill rule accepts now an optional third argument to filter for responses that were submitted not sooner than the provided value in seconds.
- 'Unknown' question types are now exported as JSON
- Study Engine:
  - `UPDATE_FLAG` action accepts other data types than strings for the value attribute.
  - `or` expression doesn't stop if any of the arguments return an error, instead it continues checking the remaining options
  - Reworked reporting system. Previously, expressions about "reports" were not used yet. Report attribute from the participant state is removed, and a new collection `<studyKey>_reports` is added.  Study actions remove due to this change: `ADD_REPORT, REMOVE_ALL_REPORTS, REMOVE_REPORT_BY_KEY, REMOVE_REPORTS_BY_KEY`.

- Study stats contain count of temporary participants as well.

### Fixed

- The "long" CSV export format now correctly displays the `metaDisplayed`, `metaResponse` and `metaVersion` columns based on their respective options, rather than based on the `metaInit` option.

## [v0.16.3] - 2022-02-15

### Changed

- Add option to set max grpc message size through environment variable

## [v0.16.2] - 2021-06-16

### Changed

- Fix issue about reference type values in study actions. Previously, change could not be detected, since old value was overwritten.

## [v0.16.1] - 2021-06-09

### Changed

- Fix issue regarding export of likert scale groups and more generally, questions that include multiple groups, and response is not found on the first level.

## [v0.16.0] - 2021-05-27

### Added

- New package: `exporter`, with main responsibility to process survey definition and responses provide them for export as CSV or survey info preview data object.
- New endpoints added:
  - `GetResponsesWideFormatCSV`: get response export in a column-wise wide format
  - `GetResponsesLongFormatCSV`: get response export in a row-wise long format
  - `GetSurveyInfoPreviewCSV`: get survey excerpt as a CSV file
  - `GetSurveyInfoPreview`: get survey excerpt as a nested data format
- New study expression: `countResponseItems`, to retrieve the numbers of how many items are in a response group.

### Changed

- Fix tests for `lastSubmissionDateOlderThan`.

## [v0.15.1] - 2021-05-24

### Changed

- `lastSubmissionDateOlderThan`: change how first argument is used. Previously it was interpreted as delta. After this change, the first argument is resolved, and interpreted as a timestamp. If a reference from now should be used, the timestampWithOffset method can be applied.

## [v0.15.0]

### Added

- New study actions:
  - `IF`: improved method for control flows with if-else logic
  - `DO`: to perform a list of actions
- Some documentation about study actions [here](docs/studyActions.md).
- Some documentation about study expression [here](docs/studyExpressions.md).
- New study expression: `checkConditionForOldResponses`
- New method in expression eval context: `mustGetStrValue` to retrieve an argument as string or produce an error.
- New gRPC endpoint for `RunRules`, to trigger custom study rules.

### Changed

- Updated dependencies

## [v0.14.1]

### Changed

- Changed how survey version ID is generated, to use a shorter format (YY-MM-counter).

### Fixed

- `HasParticipantStateWithCondition` should handle case without condition as well. (When condition is nil)

## [v0.14.0]

### Added

#### New expressions

- `getStudyEntryTime`: method to retrieve timestamp of the event, when the participant entered the study from the participant state.
- `hasSurveyKeyAssigned`: accepts one string argument with the survey key to be checked for. Returns true if the survey key exists in the assigned surveys array.
- `getSurveyKeyAssignedFrom`:  accepts one string argument with the survey key to be checked for. Returns the timestamp for the survey's validFrom attribute or -1 if the survey key is not assigned.
- `getSurveyKeyAssignedUntil`: accepts one string argument with the survey key to be checked for. Returns the timestamp for the survey's validUntil attribute or -1 if the survey key is not assigned.
- `responseHasOnlyKeysOtherThan`: expression to check if the response for a specific survey item's response group only inlcudes other keys then provided here. (E.g., symptom response contains any selection other than "no symptoms".) Returns false if response is not present at all.)
- `hasParticipantFlag`: expression to check if the participant has a specific flag. Needs two arguments for "key" and "value". Return true if key exists and value is the same as the provided second argument. Arguments can be both strings or expressions that return a string.
- `getResponseValueAsNum`: retrieve the `value` attribute of a specific survey item's selected response object as a number (float64).
- `getResponseValueAsStr`: retrieve the `value` attribute of a specific survey item's selected response object as a string.

### Changed

- Expression `timestampWithOffset` accepts optional second argument as a "reference" time. If left empty, the current time will be used as a reference.
- Expression `timestampWithOffset` accepts and can resolve Expressions for both arguments.
- Survey's version ID is now generated from the current timestamp at submission (instead of a random value). With random generated ID sometimes IDs were re-used and not unique anymore. Simply encoding the current timestamp should be enough for this purpose (Needs to be only unique within the same survey's version history). Also we can save the random number generations computing in this case.
- Adding two helper methods to study engine: (not exposed to the rule resolvers currently, but used by them)
  - `findSurveyItemResponse`: to retrieve a specific response item defined by the item's key from the array of survey item responses. Produces an error if the specific item cannot be found.
  - `findResponseObject`: to retrive a specific response object from a selected survey item's response. With the correct key (e.g. `"rg.scg.1"`), the selected response object is retrieved from the nested structure. Produces an error if the object could not be found - either parent or specific item is missing.
