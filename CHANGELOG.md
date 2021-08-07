# Changelog

## [v0.17.0] - ???

### Added

- Data exporter can parse composite question titles or option labels (when text contains multiple parts).
- Data exporter logic to handle question type: "responsive single choice array"
- Participant File Upload endpoint: method to upload files for study participants

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

