# Study Actions

The purpose of this document is to describe the currently available study actions.

### IF
`
IF( condtion, actionForTrue, actionForFalse? )
`

Helpful method for control flow, implementing the typical if-else structure.

Arguments:

1. condition, evaluted to a boolean value
2. perform this action if true
3. (optional) perform this action otherwise


### DO

`DO( action1, action2, ... )`

Method that can be used to group actions together. The `DO` method will simply iterate through the list of actions defined in the argument list.
