# Changelog

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

