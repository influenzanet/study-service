package models

import api "github.com/influenzanet/study-service/api"

// ItemComponent
type ItemComponent struct {
	Role             string            `bson:"role"`
	Key              string            `bson:"key"`
	Content          []LocalisedObject `bson:"content"`
	DisplayCondition Expression        `bson:"displayCondition"`
	Disabled         Expression        `bson:"disabled"`

	// group component
	Items []ItemComponent `bson:"items,omitempty"`
	Order *Expression     `bson:"order,omitempty"`

	// response compontent
	Dtype string `bson:"dtype,omitempty"`
}

func (comp ItemComponent) ToAPI() *api.ItemComponent {
	content := make([]*api.LocalisedObject, len(comp.Content))
	for i, si := range comp.Content {
		content[i] = si.ToAPI()
	}
	items := make([]*api.ItemComponent, len(comp.Items))
	for i, si := range comp.Items {
		items[i] = si.ToAPI()
	}

	apiComp := &api.ItemComponent{
		Role:             comp.Role,
		Key:              comp.Key,
		Content:          content,
		DisplayCondition: comp.DisplayCondition.ToAPI(),
		Disabled:         comp.Disabled.ToAPI(),
		Items:            items,

		Dtype: comp.Dtype,
	}
	if comp.Order != nil {
		apiComp.Order = comp.Order.ToAPI()
	}
	return apiComp
}

func ItemComponentFromAPI(comp *api.ItemComponent) ItemComponent {
	if comp == nil {
		return ItemComponent{}
	}
	content := make([]LocalisedObject, len(comp.Content))
	for i, si := range comp.Content {
		content[i] = LocalisedObjectFromAPI(si)
	}
	items := make([]ItemComponent, len(comp.Items))
	for i, si := range comp.Items {
		items[i] = ItemComponentFromAPI(si)
	}

	dbComp := ItemComponent{
		Role:             comp.Role,
		Key:              comp.Key,
		Content:          content,
		DisplayCondition: ExpressionFromAPI(comp.DisplayCondition),
		Disabled:         ExpressionFromAPI(comp.Disabled),
		Items:            items,

		Dtype: comp.Dtype,
	}
	if comp.Order != nil {
		exp := ExpressionFromAPI(comp.Order)
		dbComp.Order = &exp
	}
	return dbComp
}

// LocalisedObject
type LocalisedObject struct {
	Code string `bson:"code"`
	// For texts
	Parts []ExpressionArg `bson:"parts"`
}

func (o LocalisedObject) ToAPI() *api.LocalisedObject {
	parts := make([]*api.ExpressionArg, len(o.Parts))
	for i, si := range o.Parts {
		parts[i] = si.ToAPI()
	}
	return &api.LocalisedObject{
		Code:  o.Code,
		Parts: parts,
	}
}

func LocalisedObjectFromAPI(o *api.LocalisedObject) LocalisedObject {
	if o == nil {
		return LocalisedObject{}
	}
	parts := make([]ExpressionArg, len(o.Parts))
	for i, si := range o.Parts {
		parts[i] = ExpressionArgFromAPI(si)
	}

	return LocalisedObject{
		Code:  o.Code,
		Parts: parts,
	}
}
