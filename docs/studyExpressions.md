# Study Expressions

This document describes the current possibilities to evaluate study expressions.

 Study expressions are analysed and evaluated by methods of a structure that comprises the following attributes:

* `ParticipantState`: the current state of the participant in the study, i.e. which surveys are activated or which flags are set,
* `Event`: the latest event that was executed by the participant,
* `DbService`: the reference to the database abstraction layer in order to get access to previous responses of the participant.

For each expression type :

- `Functional description` describes in a function call-like syntax an overview of the parameters expected
- `Go Implementation` describes the current go function definition. Each expression is defined as a function accepting one expression (of type `types.Expression`) embedding the arguments for the call.

## 1.  countResponseItems

Counts the number of selected reponse items of a response group.

Functional Description:
```
    countResponseItems(survey, rg): float
```

Go Implementation:
```go
countResponseItems(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required:**
    
>   `expression.Data[0]` : key of the survey item convertible to type `string` \
>   `expression.Data[1]` : key of the response group convertible to type `string`


 **Note:** 
This method returns the number of selected response items of the response slot having response key `expression.Data[1]` (e.g., for a multiple choice group question) from the survey question with item key `expression.Data[0]`. If the survey question item or response group are not found, it returns -1. The length of `expression.Data` must be 2.

**Return:** `(float64, error)`

## 2. checkEventType

Checks if the latest event is of the same type as specified in the parameter expression.

Functional Description:
```
    countResponseItems(event_type): bool
```

Go Implementation:
```go
checkEventType(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**
    
>   `expression.Data[0]` : event of interest convertible to `string` 


 **Note:** 
This method checks if the latest event is of the specified type. Types of events can be e.g. "SUBMISSION", "TIMER" or "ENTER". The length of `expression.Data` must be 1.

**Return:** `(bool, error)`


<!--- Add error handling?: if not defined or error, returns false with error type-->


## 3. checkSurveyResponseKey

Checks if the key of the survey submitted during `Event` is the same as in expression data.

Functional Description:
```
    checkSurveyResponseKey(survey): bool
```

Go Implementation:
```go
checkSurveyResponseKey(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**
    
>   `expression.Data[0]` : survey key, convertible to `string` 

 **Note:** 
This method checks if the value of `expression.Data[0]` converted to `string` is equal to the key of the submitted survey provided that this key is available. Otherwise it returns `false`. The length of `expression.Data` must be 1.

**Return:** `(bool, error)`

## 4. responseHasKeysAny

checks if the participant has selected any of the specified item keys.

Functional Description:
```
    responseHasKeysAny(survey, rg_prefix, comp_key...): bool
```

Go Implementation:
```go
responseHasKeysAny(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**
    
-   `expression.Data[0]` : key of the survey item convertible to type `string` 
-   `expression.Data[1]` : the key of the response group convertible to type `string` 
-   `expression.Data[2:]` : the keys of the target answer items convertible to type `string` 

 **Note:** 
This method checks if the participant has selected any of the specified item keys `expression.Data[2:]` of the response group with response key `expression.Data[1]` from the survey question with survey item key `expression.Data[0]`. If the survey item or response object are not found, it returns `false`. The length of `expression.Data` must be at least `3`. 

**Return:** `(bool, error)`


## 5. responseHasOnlyKeysOtherThan

checks if the participant has selected none of the specified item keys.

Functional Description:
```
responseHasOnlyKeysOtherThan(survey, rg_prefix, comp_key...): bool
```

Go Implementation:
```go
responseHasOnlyKeysOtherThan(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**
    
-   `expression.Data[0]` : key of the survey item convertible to type `string` 
-   `expression.Data[1]` : key of the response group convertible to type `string` 
-   `expression.Data[2:]` : keys of the target answer items convertible to type `string` 

**Note:** This method checks if the participant has selected none of the specified item keys `expression.Data[2:]` of response group with response key `expression.Data[1]` from the survey question with survey item key `expression.Data[0]`. If the survey item or response object are not found, it returns `false`. The length of `expression.Data` must be at least `3`. 

**Return:** `(bool, error)`

## 6. getResponseValueAsNum

returns the entered numerical value of the specified response group item.

Functional Description:
```
getResponseValueAsNum(survey, rg): float
```

Go Implementation:

```go
getResponseValueAsNum(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**
    
-   `expression.Data[0]` : key of the survey item convertible to type `string` \
-   `expression.Data[1]` : key of the response group convertible to type `string` 

**Note:** This method returns the entered numerical value of the response group with response key `expression.Data[1]` (e.g. numerical input field or slider scale) from the survey question with survey item key `expression.Data[0]`. If the survey item or response object are not found, it returns `0`. The length of `expression.Data` must be `2`.

**Return:**  `(val float64, err error)`


## 7. getResponseValueAsStr

returns the entered string value of the specified response group item.

Functional Description:
```
getResponseValueAsStr(survey, rg): float
```

Go Implementation:
```go
getResponseValueAsStr(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` :  key of the survey item convertible to type `string` \
>   `expression.Data[1]` :  key of the response group convertible to type `string` 

**Note:** This method returns the entered string value of the response group with response key `expression.Data[1]` (e.g. free input fields) from the survey question with survey item key `expression.Data[0]`. If the survey item or response object are not found, it returns an empty string. The length of `expression.Data` must be `2`.


**Return:**  `(val string, err error)`


 <!--- ## 8. checkConditionForOldResponses

```go
checkConditionForOldResponses( condition, checkFor?, surveyKey?, responsesFrom?, responsesUntil? )
```

This Method runs the evaluation on old responses. An active DB connection is required for this method.

Arguments:
1. condition: [expression] - expression that should return a boolean value. The local context will contain the old survey, so it can be evaluated, as it would be submitted right now.
2. checkFor: [string/number] optional - "all"/"any" or a number
    - "all": all the found responses have to fulfil the condition, then it return true. Otherwise, or if no responses found, returns false. (Default behaviour)
    - "any": at least one of the responses has to fulfil the condition to return true. Otherwise, or if no responses found, return false.
    - number > 0: at least that many responses have to fulfil the condition to return true. Otherwise, or if no responses found, return false.
3. surveyKey: [string] optional - should return a string representing, which survey's responses should be looked up. If empty string (""), it will be ignored.
4. responsesFrom: [number] optional - filter for responses that were submitted after this timestamo. If zero, it will be ignored.
5. responsesUntil: [number] optional - filter for responses that were submitted before this timestamp. If zero, it will be ignored.

-->

## 8. checkConditionForOldResponses

Evaluates the specified expression on old responses.

Functional Description:
```
checkConditionForOldResponses(condtion[, check][, survey][,since][,until]): bool
```

Go Implementation:
```go
checkConditionForOldResponses(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
> 1. `expression.Data[0]`: an expression that should return a boolean value. As the local context will contain the old survey, it will be evaluated as it would be submitted right now.
> 2. `expression.Data[1]`: optional value of type  `string` or `float64` expecting to be `"all"`, `"any"` or a postive number:
>       * `"all"`: the method returns true if all the found responses fulfil the condition `expression.Data[0]`. Otherwise or if no responses were found, the method returns false (Default behaviour). 
>       * `"any"`: the method returns true if at least one of the responses fulfils the condition. Otherwise or if no responses  were found it returns false.
>    * `[float64] > 0`: the method returns true if at least the specified number of responses fulfils the condition. Otherwise or if no responses  were found it returns false.
> 3. `expression.Data[2]`: optional parameter of type `string` value that should contain the survey key for the responses that will be looked up. If it is an empty string (""), it will be ignored.
> 4. `expression.Data[3]`:  optional number parameter that will be interpreted as timestamp to filter for responses that were submitted after this date. In case of zero it will be ignored.
> 5. `expression.Data[4]`:  optional number parameter that will be interpreted as timestamp to filter for responses that were submitted before this date. In case of zero it will be ignored.


**Note:**  An active DB connection is required for this method.

**Return:**  `(bool, error)`

## 9. getStudyEntryTime

Returns the time (as posix timestamp) the participant entered the study.

Functional Description:
```
getStudyEntryTime(): float
```

Go Implementation:
```go
getStudyEntryTime(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Return:**  `(float64, error)`

## 10. hasSurveyKeyAssigned

Checks if the specified survey key is included in the keys of the surveys assigned to the participant.

Functional Description:
```
hasSurveyKeyAssigned(survey): bool
```

Go Implementation:

```go
hasSurveyKeyAssigned(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` : survey key of interest as type `string`. 

**Note:** The length of `expression.Data` must be `1`.

**Return:**  `(val bool, err error)`


## 11. getSurveyKeyAssignedFrom

Returns the date when the specified survey was assigned to the participant as posix timestamp.

Functional Description:
```
getSurveyKeyAssignedFrom(survey): float
```

Go Implementation:
```go
getSurveyKeyAssignedFrom(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` : should contain the survey key of interest as type `string`. 

**Note:** If none of the assigned surveys has the specified survey key, it returns -1. The length of `expression.Data` must be `1`. 

**Return:**  `(float64, error)`


## 12. getSurveyKeyAssignedUntil

Returns the date until the specified survey should be submitted by the participant as posix timestamp.

Functional Description:
```
getSurveyKeyAssignedUntil(survey): float
```

Go Implementation:

```go
getSurveyKeyAssignedUntil(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` : should contain the survey key of interest as type `string`. 


**Note:** If none of the assigned surveys has the specified survey key, it returns `-1`. The length of `expression.Data` must be `1`.

**Return:**  `(float64, error)`

## 13. hasStudyStatus

Checks if the participant has the specified status.

Functional Description:
```
hasStudyStatus(status): bool
```

Go Implementation:
```go
hasStudyStatus(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` : should contain the status value of type `string` that is compared to the current status of the participant. 

**Note:** Possible values for the status of the paticipant are `"active"`, `"inactive"`, `"paused"`, `"finished"`. The length of `expression.Data` must be `1`.

**Return:**  `(bool, error)`

## 14. hasParticipantFlag

Checks if the participant has the specified flag set with a given value.

Functional Description:
```
hasParticipantFlag(flag_key, value): bool
```

Go Implementation:

```go
hasParticipantFlag(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` : should contain the name of the flag of interest as `string` \
>   `expression.Data[1]` : should contain the flag value of interest as `string`.

**Note:** The length of `expression.Data` must be `2`.

**Return:**  `(bool, error)`


## 15. lastSubmissionDateOlderThan

Checks if the submission date either of the last survey submitted or the specified survey is older than the specified date. 

Functional Description:
```
lastSubmissionDateOlderThan(time[, survey]): bool
```

Go Implementation:
```go
lastSubmissionDateOlderThan(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  
>   `expression.Data[0]` : should contain the date of interest as posix time stamp. \
>   `expression.Data[1]` : optional parameter that should contain the key of the survey whose submission date will be compared to the first argument `expression.Data[0]`.

**Note:** The length of `expression.Data` must be `1` or `2`.

**Return:**  `(bool, error)`


## 16. eq

Checks if the first two entries of expression data are equal.

Functional Description:
```
eq(v1, v2): bool
```

Go Implementation:
```go
eq(expression) 
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**
    
>   `expression.Data[0]` : should be a value of type `string` or `float64` \
>   `expression.Data[1]` : should be a value of type `string` or `float64`


 **Note:**  The type of the arguments should be either both `string` or `float64`. The length of `expression.Data` must be 2.

**Return:** `(bool, error)`


## 17. lt

Checks if the first entry of expression data is less than the second entry.

Functional Description:
```
lt(v1, v2): bool
```

Go Implementation:
```go
lt(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  

>   `expression.Data[0]` : should be a value of type `string` or `float64` \
>   `expression.Data[1]` : should be a value of type `string` or `float64`


 **Note:** Strings are compared lexicographically. The type of the arguments should be either both `string` or `float64`. The length of `expression.Data` must be 2.

**Return:** `(bool, error)`


## 18. lte

Checks if the first entry of expression data is less than or equal to the second entry.

Functional Description:
```
lte(v1, v2): bool
```

Go Implementation:
```go
lte(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  

>   `expression.Data[0]` : should be a value of type `string` or `float64` \
>   `expression.Data[1]` : should be a value of type `string` or `float64`


 **Note:** Strings are compared lexicographically. The type of the arguments should be either both `string` or `float64`. The length of `expression.Data` must be 2.

**Return:** `(bool, error)`

## 19. gt

Checks if the first entry of expression data is greater than the second entry.

Functional Description:
```
gt(v1, v2): bool
```

Go Implementation:

```go
gt(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  

>   `expression.Data[0]` : should be a value of type `string` or `float64` \
>   `expression.Data[1]` : should be a value of type `string` or `float64`


 **Note:** Strings are compared lexicographically. The type of the arguments should be either both `string` or `float64`. The length of `expression.Data` must be 2.

**Return:** `(bool, error)`

## 20. gte

Checks if the first entry of expression data is greater than or equal to the second entry.

Functional Description:
```
gte(v1, v2): bool
```

Go Implementation:

```go
gte(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  

>   `expression.Data[0]` : should be a value of type `string` or `float64` \
>   `expression.Data[1]` : should be a value of type `string` or `float64`


 **Note:** Strings are compared lexicographically. The type of the arguments should be either both `string` or `float64`. The length of `expression.Data` must be 2.


## 22. and

Checks if all entries of expression data are unequal to zero or `true`.

Functional Description:
```
and(expressions...): bool
```

Go Implementation:
```go
and(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  

>   `expression.Data[:]` : only values of type `bool` or `float64` are evaluated

**Note:**
The length of `expression.Data` must be at least `2`. This Method returns `true` if and only if
* all of the arguments of type `bool` are `true` and
* all of the arguments of type `float64` have a value that is not zero.

**Return:** `(bool, error)`


## 23. or

Checks if there is one entry of expression data that is `true`or greater than zero.

Functional Description:
```
or(expressions...): bool
```

Go Implementation:
```go
or(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required in this method:**  

>   `expression.Data[:]` : only values of type `bool` or `float64` are evaluated


**Note:** The length of `expression.Data` must be at least `2`. This Method returns `true` if 
* one of the arguments of type `bool` is `true` or
* one of the arguments of type `float64` has a value greater than zero.

**Return:** `(bool, error)`

## 24. not

Checks if the first entry of expression data is `0` or `false`.

Functional Description:
```
    not(value): bool
```

Go Implementation:
```go
not(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required:**  

>   `expression.Data[0]` : should be a value of type `bool` or `float64`

**Note:** The length of `expression.Data` must be 1. The Method returns `true` if and only if either
* `expression.Data[0]` is type `bool` and `false`  or 
* `expression.Data[0]` is type `float64` and `0`. 

**Return:** `(bool, error)`

## 25. timestampWithOffset

Returns the specified offset time added to either the current time or the specified reference time.

Functional Description:
```
timestampWithOffset(offset[, ref_time]: float
```

Go Implementation:
```go
timestampWithOffset(expression)
```

**Parameter:**    `expression` : a structure of `types.Expression` 

**Required:**  
>   `expression.Data[0]` : should contain the offset time convertible to `int64` \
>   `expression.Data[1]` : optional parameter that should contain the reference time convertible to `int64`.

**Return:**  `(float64, error)`
