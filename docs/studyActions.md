# Study Actions

This document describes the currently available study actions provided by CASE.

 Study actions are performed by using the following collection of `struct` objects as parameters:

* `action` : an object of `types.Expression`. Specifies the actions that should be performed.
* `oldState` : an object of `types.ParticipantState`. Characterizes the current state of the participant that will potentially be updated by evaluating the action expression.
* `event` : an object of `types.StudyEvent`. Specifies the event that was triggered by the participant or by the program (e.g. `"TIMER"`) and collects the survey responses,
* `dbService` : an object of `types.StudyDBService`. References the database abstraction layer to get access to previous responses of the participant (for example to check them for conditions specified by the researcher).

The functions executing actions are listed in the following.
The header denotes the string keyword leading to the decision which kind of action will be performed. The block code indicates the header of the function that will be executed in case of the keyword specified.

## 1.  IF

This function is used for control flow implementing the typical if-else structure. Conditions are checked in order to decide which actions will be performed.

Functional description:
```
IF(condition, action[, action_else])
```

Go Implementation:
```go
ifAction(action, oldState, event, dbService)
```

**Required Parameter:**

>   `action.Data[0]` :  condition, evaluated to a boolean value \
>   `action.Data[1]` : perform this action if condition is true  \
>   `action.Data[2]` :  (optional) perform this action otherwise


 **Note:** If the execution of action fails, the old state of the participant is returned unmodified. The length of `action.Data` must be at least 2.

**Return:** `(types.ParticipantState, error)`


### 2. DO

Performs a list of actions by iterating through the `action.Data` argument. This function can be used to group actions together as a defined list of arguments.

Functional description:
```
DO(action...)
```

Go Implementation:
```go
doAction(action, oldState, event, dbService)
```

**Required Parameters:**

>   `action.Data[0:]` :  list of actions that will be performed successively

**Return:** `(types.ParticipantState, error)`


### 3. IFTHEN

Conditionally performs a list of actions. The function checks the first argument as condition deciding if actions will be performed. In case of `true` it iterates through the following entries arguments.

Functional description:
```
IFTHEN(condition, actions...)
```

Go Implementation:
```go
ifThenAction(action, oldState, event, dbService)
```

**Required Parameters:**

>   `action.Data[0]` :  condition, evaluated to a boolean value \
>   `action.Data[1:]` : perform this sequence of actions if condition is true

 **Note:** The length of `action.Data` must be at least 1.

**Return:** `(types.ParticipantState, error)`


## 4. UPDATE_STUDY_STATUS

Updates the status of the participant (e.g. from active to inactive). Possible status values: "active", "temporary", "exited". Other values are possible and are handled like "exited" on the server.

Functional description:
```
UPDATE_STUDY_STATUS(status)
```

Go Implementation:
```go
updateStudyStatusAction(action, oldState, event)
```


**Required Parameter:**

>   `action.Data[0]` : the new status of the participant convertible to `string`. Possible values are: `"active"`, `"temporary"`, `"exited"`. Other values are possible and are handled like `"exited"` on the server.


 **Note:**
 The length of `action.Data` must be 1.

**Return:** `(types.ParticipantState, error)`

## 5. START_NEW_STUDY_SESSION

Generates an ID for a new study session.		
    
Functional description:
```
START_NEW_STUDY_SESSION()
```

Go Implementation:
```go
startNewStudySession(action, oldState, event)
```
	

**Return:** `(types.ParticipantState, error)`
		


## 6. UPDATE_FLAG

Updates one flag of the participant state. The flag attribute of the `ParticipantState` object is a map with string keys addressing corresponding string values.

Functional description:
```
UPDATE_FLAG(flag_key, value)
```

Go Implementation:
```go
updateFlagAction(action, oldState, event)
```

**Required Parameters:**

>   `action.Data[0]` : the string key of the flag to be updated \
>   `action.Data[1]` : the string value of the flag to be updated

 **Note:**
 The length of `action.Data` must be 2.

**Return:** `(types.ParticipantState, error)`

## 7. REMOVE_FLAG

Deletes the flag with the specified key of the participant state. The flag attribute of the `ParticipantState` object is a map with string keys addressing corresponding string values.

Functional description:
```
REMOVE_FLAG(flag_key)
```

Go Implementation:
```go
removeFlagAction(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the string key of the flag to be removed


 **Note:**
 The length of `action.Data` must be 1.

**Return:** `(types.ParticipantState, error)`


## 8. ADD_NEW_SURVEY

Appends a new survey to the assigned surveys of the participant state (expressed by the attribute `AssignedSurveys` of `ParticipantState`).

Functional description:
```
  ADD_NEW_SURVEY(survey, since, until, mode )
```

Go Implementation:
```go
addNewSurveyAction(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the string key of the survey to be assigned to the participant \
>   `action.Data[1]` : a float value indicating the timestamp from which the assigned survey is visible \
>   `action.Data[2]` : a float value indicating the time until the assigned survey is visible and should be submitted \
>   `action.Data[3]` : a string value indicating the mode of displaying the assigned survey. Possible values are `"prio"`, `"normal"`, `"quick"` or `"update"`.


 **Note:**
 The length of `action.Data` must be 4.

**Return:** `(types.ParticipantState, error)`


## 9. REMOVE_ALL_SURVEYS

Clears the list of assigned surveys of participant state.

Functional description:
```
  REMOVE_ALL_SURVEYS()
```

Go Implementation:
```go
removeAllSurveys(action, oldState, event)
```

 **Note:**
 Arguments of `action.Data` are permitted for this function.

**Return:** `(types.ParticipantState, error)`


## 10. REMOVE_SURVEY_BY_KEY

Removes the first or last occurence of a survey with specific key in the list of assigned surveys of the participant state.

Functional description:
```
  REMOVE_SURVEY_BY_KEY(survey, position)
```

Go Implementation:
```go
removeSurveyByKey(action, oldState, event)
```

**Required Parameters :**

>   `action.Data[0]` : the string key of the survey to be removed at first or last occurence. \
>   `action.Data[1]` : a string value indicating if the first or last occurence of an asssigned survey should be removed. Expected values are `"first"` or `"last"`.

**Return:** `(types.ParticipantState, error)`

## 11. REMOVE_SURVEYS_BY_KEY

Removes all surveys with the specified key in the list of assigned surveys of the participant state.

Functional description:
```
  REMOVE_SURVEYS_BY_KEY(survey)
```

Go Implementation:

```go
removeSurveysByKey(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the string key of the surveys to be removed.


**Return:** `(types.ParticipantState, error)`


## 12. ADD_MESSAGE

appends a message to the message array of participant state.

Functional description:
```
ADD_MESSAGE(messageType, timestamp)
```

Go Implementation:

```go
addMessage(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the message type as string that specifies which template message should be send.
>   `action.Data[1]` : the timestamp as float the message is scheduled for.

 **Note:**
 The length of `action.Data` must be 2.

**Return:** `(types.ParticipantState, error)`

## 13. REMOVE_ALL_MESSAGES

Clears the message list of participant state.

Functional description:
```
  REMOVE_ALL_MESSAGES()
```

Go Implementation:
```go
removeAllMessages(action, oldState, event)
```

 **Note:**  No arguments are allowed.

**Return:** `(types.ParticipantState, error)`


## 13. REMOVE_MESSAGES_BY_TYPE

Removes all messages with the specified type in the list of messages of the participant state.

Functional description:
```
  REMOVE_MESSAGES_BY_TYPE(messageType)
```

Go Implementation:
```go
removeMessagesByType(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the string type of the messages to be removed. \


 **Note:**
 The length of `action.Data` must be 1.

**Return:** `(types.ParticipantState, error)`



## 13. REMOVE_REPORT_BY_KEY

Removes the first or last occurence of a survey item response with specific key in the list of reports of the participant state.

Functional description:
```
  REMOVE_REPORT_BY_KEY(survey, position)
```

Go Implementation:
```go
removeReportByKey(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the string key of the survey item response to be removed at first or last occurence in reports. \
>   `action.Data[1]` : a string value indicating if the first or last occurence of an survey item response should be removed. Expected values are `"first"` or `"last"`.


 **Note:**
 The length of `action.Data` must be 2.

**Return:** `(types.ParticipantState, error)`


## 14. REMOVE_REPORTS_BY_KEY

Removes all survey item responses with the specified key in the list of reports of the participant state.

Functional description:
```
  REMOVE_REPORTS_BY_KEY(survey)
```

Go Implementation:
```go
removeReportsByKey(action, oldState, event)
```

**Required Parameter:**

>   `action.Data[0]` : the string key of the survey item reponses to be removed.

**Return:** `(types.ParticipantState, error)`
