package models

import api "github.com/influenzanet/study-service/api"

// ItemComponent
type ItemComponent struct {
	Role             string            `bson:"role"`
	Key              string            `bson:"key"`
	Content          []LocalisedObject `bson:"content"`
	DisplayCondition *Expression       `bson:"displayCondition"`
	Disabled         *Expression       `bson:"disabled"`

	// group component
	Items []ItemComponent `bson:"items,omitempty"`
	Order *Expression     `bson:"order,omitempty"`

	// response compontent
	Dtype      string               `bson:"dtype,omitempty"`
	Properties *ComponentProperties `bson:"properties,omitempty"`

	Style       []Style           `bson:"style,omitempty"`
	Description []LocalisedObject `bson:"description"`
}

func (comp *ItemComponent) ToAPI() *api.ItemComponent {
	if comp == nil {
		return nil
	}
	content := make([]*api.LocalisedObject, len(comp.Content))
	for i, si := range comp.Content {
		content[i] = si.ToAPI()
	}
	description := make([]*api.LocalisedObject, len(comp.Description))
	for i, si := range comp.Description {
		description[i] = si.ToAPI()
	}

	items := make([]*api.ItemComponent, len(comp.Items))
	for i, si := range comp.Items {
		items[i] = si.ToAPI()
	}
	style := make([]*api.ItemComponent_Style, len(comp.Style))
	for i, si := range comp.Style {
		style[i] = si.ToAPI()
	}

	apiComp := &api.ItemComponent{
		Role:             comp.Role,
		Key:              comp.Key,
		Content:          content,
		DisplayCondition: comp.DisplayCondition.ToAPI(),
		Disabled:         comp.Disabled.ToAPI(),
		Items:            items,

		Dtype:       comp.Dtype,
		Properties:  comp.Properties.ToAPI(),
		Style:       style,
		Description: description,
	}
	if comp.Order != nil {
		apiComp.Order = comp.Order.ToAPI()
	}
	return apiComp
}

func ItemComponentFromAPI(comp *api.ItemComponent) *ItemComponent {
	if comp == nil {
		return nil
	}
	content := make([]LocalisedObject, len(comp.Content))
	for i, si := range comp.Content {
		content[i] = LocalisedObjectFromAPI(si)
	}
	description := make([]LocalisedObject, len(comp.Description))
	for i, si := range comp.Description {
		description[i] = LocalisedObjectFromAPI(si)
	}
	items := make([]ItemComponent, len(comp.Items))
	for i, si := range comp.Items {
		items[i] = *ItemComponentFromAPI(si)
	}

	style := make([]Style, len(comp.Style))
	for i, si := range comp.Style {
		style[i] = StyleFromAPI(si)
	}

	dbComp := ItemComponent{
		Role:             comp.Role,
		Key:              comp.Key,
		Content:          content,
		DisplayCondition: ExpressionFromAPI(comp.DisplayCondition),
		Disabled:         ExpressionFromAPI(comp.Disabled),
		Items:            items,

		Dtype:       comp.Dtype,
		Properties:  ComponentPropertiesFromAPI(comp.Properties),
		Style:       style,
		Description: description,
	}
	if comp.Order != nil {
		exp := ExpressionFromAPI(comp.Order)
		dbComp.Order = exp
	}
	return &dbComp
}

type Style struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
}

func (s Style) ToAPI() *api.ItemComponent_Style {
	return &api.ItemComponent_Style{
		Key:   s.Key,
		Value: s.Value,
	}
}

func StyleFromAPI(st *api.ItemComponent_Style) Style {
	if st == nil {
		return Style{}
	}
	return Style{
		Key:   st.Key,
		Value: st.Value,
	}
}

type ComponentProperties struct {
	Min           *ExpressionArg `bson:"min"`
	Max           *ExpressionArg `bson:"max"`
	StepSize      *ExpressionArg `bson:"stepSize"`
	DateInputMode *ExpressionArg `bson:"dateInputMode"`
}

func (s *ComponentProperties) ToAPI() *api.ItemComponent_Properties {
	if s == nil {
		return nil
	}
	return &api.ItemComponent_Properties{
		Min:           s.Min.ToAPI(),
		Max:           s.Max.ToAPI(),
		StepSize:      s.StepSize.ToAPI(),
		DateInputMode: s.DateInputMode.ToAPI(),
	}
}

func ComponentPropertiesFromAPI(st *api.ItemComponent_Properties) *ComponentProperties {
	if st == nil {
		return nil
	}
	return &ComponentProperties{
		Min:           ExpressionArgFromAPI(st.Min),
		Max:           ExpressionArgFromAPI(st.Max),
		StepSize:      ExpressionArgFromAPI(st.StepSize),
		DateInputMode: ExpressionArgFromAPI(st.DateInputMode),
	}
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
		parts[i] = *ExpressionArgFromAPI(si)
	}

	return LocalisedObject{
		Code:  o.Code,
		Parts: parts,
	}
}
