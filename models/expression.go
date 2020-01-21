package models

import (
	"log"

	api "github.com/influenzanet/study-service/api"
)

type Expression struct {
	Name  string          `bson:"name"`
	DType string          `bson:"dtype"`
	Data  []ExpressionArg `bson:"data"`
}

type ExpressionArg struct {
	Exp Expression `bson:"exp"`
	Str string     `bson:"str"`
	Num float64    `bson:"num"`
}

func (e ExpressionArg) ToAPI() *api.ExpressionArg {
	eargs := &api.ExpressionArg{}
	if len(e.Exp.Name) > 0 {
		eargs.Data = &api.ExpressionArg_Exp{Exp: e.Exp.ToAPI()}
	} else if len(e.Str) > 0 {
		eargs.Data = &api.ExpressionArg_Str{Str: e.Str}
	} else {
		eargs.Data = &api.ExpressionArg_Num{Num: e.Num}
	}
	return eargs
}

func (e Expression) ToAPI() *api.Expression {
	data := make([]*api.ExpressionArg, len(e.Data))
	for i, ea := range e.Data {
		data[i] = ea.ToAPI()
	}
	return &api.Expression{
		Name:  e.Name,
		Dtype: e.DType,
		Data:  data,
	}
}

func ExpressionArgFromAPI(e *api.ExpressionArg) ExpressionArg {
	newEA := ExpressionArg{}
	if e == nil {
		return newEA
	}

	switch x := e.Data.(type) {
	case *api.ExpressionArg_Exp:
		newEA.Exp = ExpressionFromAPI(x.Exp)
	case *api.ExpressionArg_Str:
		newEA.Str = x.Str
	case *api.ExpressionArg_Num:
		newEA.Num = x.Num
	case nil:
		// The field is not set.
	default:
		log.Printf("api.ExpressionArg has unexpected type %T", x)
	}
	return ExpressionArg{}
}

func ExpressionFromAPI(e *api.Expression) Expression {
	exp := Expression{}
	if e == nil {
		return exp
	}
	exp.Name = e.Name
	exp.DType = e.Dtype

	exp.Data = make([]ExpressionArg, len(e.Data))
	for i, ea := range e.Data {
		exp.Data[i] = ExpressionArgFromAPI(ea)
	}
	return exp
}
