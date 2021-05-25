# Study Expressions

### checkConditionForOldResponses

`
checkConditionForOldResponses( condition, checkFor?, surveyKey?, responsesFrom?, responsesUntil? )
`

Method to run evaluation on old responses. An active DB connection is required for this method.

Arguments:
1. condition: [expression] - expression that should return a boolean value. The local context will contain the old survey, so it can be evaluated, as it would be submitted right now.
2. checkFor: [string/number] optional - "all"/"any" or a number
    - "all": all the found responses have to fulfil the condition, then it return true. Otherwise, or if no responses found, returns false. (Default behaviour)
    - "any": at least one of the responses has to fulfil the condition to return true. Otherwise, or if no responses found, return false.
    - number > 0: at least that many responses have to fulfil the condition to return true. Otherwise, or if no responses found, return false.
3. surveyKey: [string] optional - should return a string representing, which survey's responses should be looked up. If empty string (""), it will be ignored.
4. responsesFrom: [number] optional - filter for responses that were submitted after this timestamo. If zero, it will be ignored.
5. responsesUntil: [number] optional - filter for responses that were submitted before this timestamp. If zero, it will be ignored.