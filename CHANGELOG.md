# Changelog

## [v1.8.1] - 2025-01-14

### Changed

- exporter for responsive single choice array always using the list processing method (fix missing export, when only one row)

## [v1.8.0] - 2024-11-06

### Changed

- add json annotation for survey responses and expressions
- Allow to customize current time in studyengine using Now variable (function returing current time, time.Now() by default)
- add new Study Expressions `sum` and `neg` for arithmetic operations
- add new study expression `getLastSubmissionDate`
- add `MapToKey` in SurveyItemResponse to map confidential reponse to another key

## [v1.7.5] - 2024-08-13

### Changed

- when looking for latest survey def, unpublished might be missing

## [v1.7.4] - 2024-08-12

### Changed

- Can override lifetime of temporary participants with `TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD` environment variable. (Interpreted as seconds)
- Use study-rules collection to retrieve current study rules, but allow fallback to study info object's rules attribute if not found in the collection. Also it can parse serielized rules (rules attribute stored as string to avoid max nesting level issue in MongoDB).

## [v1.7.3] - 2024-01-12

### Changed

- `GetReportsForUser` now supports limit parameter (backward compatible)

## [v1.7.2] - 2023-10-24

### Changed

- Fix export of expression args, when dtype is set to "str" but string value is empty.

## [v1.7.1] - 2023-10-12

### Added

- The service will create a database index on all studies' participants collection in all available instances for `participantID`.

### Changed

- PR24 by dependabot bumping project dependencies

## [v1.7.0] - 2023-09-21

### Added

- New db collection for history of study rules, new methods to get, delete and add study rules.
- New uploadedAt index model for study rules
- New gRPC endpoints:
  - `GetStudyRulesHistory`: get study rules history for specified study key that fullfill query criteria with pagination and sorting options
  - `GetCurrentStudyRules`: get current study rules for specified study key
  - `RemoveStudyRulesVersion`: deletes study rules version with specified id.
  - `GetResponsesFlatJSONWithPagination`: stream responses in JSON format with pagination infos.
- New study-engine expressions
  - `getISOWeekForTs`: accepts one argument as timestamp and returns the ISO week number for the given timestamp.
  - `getTsForNextISOWeek`: first argument should be a value between 1-53 (ISO week number), second argument optionally can define a reference time (if not provided, current time is used). Returns timestamp for the beginning of the week for the first time that is later than the reference and has the ISO week as defined in argument 1.

### Changed

- Changed gRPC endpoints:
  - `DeleteStudy`: study rules history is also deleted
  - `SaveStudyRules`: new study rules are added in study rules collection

## [v1.6.2] - 2023-07-27

### Added

- New study-engine expression to generate a random number: `generateRandomNumber`, with two required arguments for the min and max value (both inclusive).

### Changed

- Improved pagination code.

## [v1.6.1] - 2023-07-13

### Changed

- Improvements on the `GetParticipantStatesWithPagination` endpoint to handle parameter validation.

## [v1.6.0] - 2023-06-15

### Added

- New gRPC endpoints:
  - `GetStudiesWithPendingParticipantMessages`: get all studies that have pending messages for a participant.
  - `GetParticipantStateByID`: get participant state with matching id.
  - `GetParticipantStatesWithPagination`: get participant states that fullfill query criteria with pagination and sorting option.
    - EXAMPLE for query parameter: query = `{"enteredAt":{"$gt":1686806848},"$and":[{"flags.ageCategory":"adult"},{"studyStatus":"active"}]}`
    - EXAMPLE for sortBy parameter: sortBy = `{"enteredAt": -1}`

### Changed

- `ensureDBindexes`: ensures that MongoDB indexes are created on the fields
  - `surveyDefinition` in the survey collection for all studies,
  - `scheduledFor` of `messages` object array in the participant collection for all studies.
- Include improved logger, using configurable log levels.
- Updated project dependencies.

## [v1.5.0] - 2023-03-20

### Added

- New api endpoint `ProfileDeleted`, that can be used to notify a study about the event that a user profile has been deleted. This will trigger the 'LEAVE' study event, mark the participant study status as 'accountDeleted' and remove all confidential data for all the studies the participant is enrolled in.

### Changed

- External event handlers and external expression eval:
  - added the possibility to define a custom timeout for the external call
  - external service can be configured to use mutual TLS authentication
  - expressions can use a second argument to define the API route (if defined, the route will be appended to the base URL)
- When participant is created during the `ENTER` study event, the `enteredAt` timestamp is now set to the middle of the day to improve privacy.
- Improvement in logging

## [v1.4.0] - 2023-02-12

- [PR15](https://github.com/influenzanet/study-service/compare/master...exporter-changes)

### Added

- New study expression `parseValueAsNum`: accepts one argument and attempts to parse the value of this resolved argument as float64. If value is already number, the value is returned. If argument is an expression, it will be first evaluated. Strings will be attempted to be parsed. Boolean value and strings that cannot be parsed as a number return an error.
- Implement new option types `OPTION_TYPE_EMDEBBED_CLOZE_XXX` for cloze options within single choice and multiple choice questions:
  - `OPTION_TYPE_EMDEBBED_CLOZE_TEXT_INPUT` for text,
  - `OPTION_TYPE_EMDEBBED_CLOZE_DATE_INPUT` for dates,
  - `OPTION_TYPE_EMDEBBED_CLOZE_NUMBER_INPUT` for numbers,
  - `OPTION_TYPE_EMDEBBED_CLOZE_DROPDOWN` for dropdown.
- Add embedded cloze input options of single choice and multiple choice questions to JSON file of survey info.
- Slots for embedded cloze option types within single choice and multiple choice questions are now always generated in data exorter regardless of the presence of answers in response.

### Changed

- ExpArg resolver, when trying to resolve an expression type, checking for nil values to prevent a crash in case the study rules contain wrong arguments.
- Ignore confidential questions in data exporter (they have separate export path)
- Update exporter documentation:
  - add info about column `session`
  - add links to sections
  - add info about responsive matrix question types
  - enhance clarity and readability of text
  - update info about column `version`
- Update Mapping of survey response to survey defintion. Mapping is performed by the following steps:
  - search for same version IDs, if not found
  - search for nearest version with published date < response submission date, with either response submission date < unpublished date or version is still published, if not found,
  - search for nearest version with published date > response submission date, if not found
  - take nearest version with published date < response submission date.

## [v1.3.1] - 2022-11-16

### Changed

- Data exporter should now correctly include the context columns when exporting as JSON file.

## [v1.3.0] - 2022-10-29

### BREAKING CHANGE

- [PR13](https://github.com/influenzanet/study-service/pull/13): reworked survey defintion data model to have a more scalable solution regarding number of versions. Instead of storing every survey version into one document, with the changes of this release, survey versions are stored into separate documents. The facilitate the interaction with the new version model, the API was also modified.
  - A migration tool is provided in [/tools/survey_histroy_model_migration](/tools/survey_histroy_model_migration) to convert existing DB collections or JSON files to use the new model.
  - API Changes:
    - `GetSurveyDefForStudy` can be used to retrieve a specific survey version.
    - Replaced `RemoveSurveyFromStudy` with `RemoveSurveyVersion` that can be used to delete a specific survey version.
    - `GetSurveyVersionInfos` is a new method to retrieve versions of a survey (without the survey content, to reduce size, use `GetSurveyDefForStudy` to get content for a specific version).
    - `GetSurveyKeys` is a new method to fetch survey keys for a study.
    - `UnpublishSurvey` is a new method to mark all existing survey versions "unpublished".

### Added

- [PR14](https://github.com/influenzanet/study-service/pull/14): possibility to encode participant ID with base64 URL encoding. This results in a shorter string compared to the hex encoding.

### Changed

- Upgrade project dependencies -> minimum go version increased to v1.17

## [v1.2.1] - 2022-10-06

### Changed

- Changed `PREFILL_SLOT_WITH_VALUE` to allow definition of multiple slot values within one question. To implement this, Response item slice has been changed to pointer type.

## [v1.2.0] - 2022-10-06

### Added

- Exporter supports `responsiveMatrix` response type.

### Changed

- Include participant flags when sending infos for participant messages, so these can be optionally used for message template execution.

## [v1.1.1] - 2022-09-28

### Changes

- This release fixes bugs that prevented deletion of confidential data with TIMER based study events and the participant state to be stored after TIMER events.

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
  - Reworked reporting system. Previously, expressions about "reports" were not used yet. Report attribute from the participant state is removed, and a new collection `<studyKey>_reports` is added. Study actions remove due to this change: `ADD_REPORT, REMOVE_ALL_REPORTS, REMOVE_REPORT_BY_KEY, REMOVE_REPORTS_BY_KEY`.

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
- `getSurveyKeyAssignedFrom`: accepts one string argument with the survey key to be checked for. Returns the timestamp for the survey's validFrom attribute or -1 if the survey key is not assigned.
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
