package study

import (
	"fmt"
	"reflect"
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
