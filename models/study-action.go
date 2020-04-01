package models

type Action struct {
	Name      string
	Args      []ExpressionArg // arguments if specific action methods
	Condition ExpressionArg   // for IFTHEN - if evaluates to true perform actions
	Actions   []Action        // for IFTHEN - perform these actions
}
