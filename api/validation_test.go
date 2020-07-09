package api

import "testing"

func TestEndpoint_validate(t *testing.T) {
	type fields struct {
		URL          string
		EndpointType string
		Identifier   Identifier
		Status       string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "ok", fields: fields{Identifier: "id", URL: "foo:bar", EndpointType: "fhir"}, wantErr: false},
		{name: "missing type", fields: fields{Identifier: "id", URL: "foo:bar"}, wantErr: true},
		{name: "missing url", fields: fields{Identifier: "id"}, wantErr: true},
		{name: "missing all", fields: fields{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Endpoint{
				URL:          tt.fields.URL,
				EndpointType: tt.fields.EndpointType,
				Identifier:   tt.fields.Identifier,
				Status:       tt.fields.Status,
			}
			if err := e.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrganization_validate(t *testing.T) {
	type fields struct {
		Endpoints  *[]Endpoint
		Identifier Identifier
		Keys       *[]JWK
		Name       string
		PublicKey  *string
	}
	keys := &[]JWK{{AdditionalProperties: map[string]interface{}{}}}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "ok", fields: fields{Identifier: "id", Name: "hello", Keys: keys}, wantErr: false},
		{name: "missing name", fields: fields{Identifier: "id"}, wantErr: true},
		{name: "missing id", fields: fields{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := Organization{
				Endpoints:  tt.fields.Endpoints,
				Identifier: tt.fields.Identifier,
				Keys:       tt.fields.Keys,
				Name:       tt.fields.Name,
				PublicKey:  tt.fields.PublicKey,
			}
			if err := o.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nonEmptyString(t *testing.T) {
	type args struct {
		value string
		name  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "ok", args: args{value: "v", name: "v"}, wantErr: false},
		{name: "empty", args: args{value: "", name: "v"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := nonEmptyString(tt.args.value, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("nonEmptyString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
