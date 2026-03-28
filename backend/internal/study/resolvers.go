package study

import (
	"fmt"
	"reflect"
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func structStringField(source interface{}, field string) string {
	v := reflect.ValueOf(source)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	f := v.FieldByName(field)
	if !f.IsValid() {
		return ""
	}
	if s, ok := f.Interface().(string); ok {
		return s
	}
	return ""
}

func stringField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		return structStringField(p.Source, field), nil
	}
}

func stringSliceField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		v := reflect.ValueOf(p.Source)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return []string{}, nil
		}
		f := v.FieldByName(field)
		if !f.IsValid() {
			return []string{}, nil
		}
		if s, ok := f.Interface().([]string); ok {
			return s, nil
		}
		return []string{}, nil
	}
}

func intField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		v := reflect.ValueOf(p.Source)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return 0, nil
		}
		f := v.FieldByName(field)
		if !f.IsValid() {
			return 0, nil
		}
		if i, ok := f.Interface().(int); ok {
			return i, nil
		}
		return 0, nil
	}
}

func floatField(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		v := reflect.ValueOf(p.Source)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return 0.0, nil
		}
		f := v.FieldByName(field)
		if !f.IsValid() {
			return 0.0, nil
		}
		if fv, ok := f.Interface().(float64); ok {
			return fv, nil
		}
		return 0.0, nil
	}
}

func oidResolver(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		source := reflect.ValueOf(p.Source)
		if source.Kind() == reflect.Ptr {
			source = source.Elem()
		}
		if source.Kind() != reflect.Struct {
			return "", nil
		}
		f := source.FieldByName(field)
		if !f.IsValid() {
			return "", nil
		}
		if oid, ok := f.Interface().(primitive.ObjectID); ok {
			if oid.IsZero() {
				return "", nil
			}
			return oid.Hex(), nil
		}
		return fmt.Sprintf("%v", f.Interface()), nil
	}
}

func timeResolver(field string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		source := reflect.ValueOf(p.Source)
		if source.Kind() == reflect.Ptr {
			source = source.Elem()
		}
		if source.Kind() != reflect.Struct {
			return time.Now().UTC().Format(time.RFC3339), nil
		}
		f := source.FieldByName(field)
		if !f.IsValid() {
			return time.Now().UTC().Format(time.RFC3339), nil
		}
		t, ok := f.Interface().(time.Time)
		if !ok {
			return time.Now().UTC().Format(time.RFC3339), nil
		}
		return t.UTC().Format(time.RFC3339), nil
	}
}
