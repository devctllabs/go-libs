package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
)

//go:generate go tool mockgen -destination mocks/config.gen.go -package mocks . Input,Loader

// Loader applies one configuration source to a caller-owned target.
type Loader interface {
	// Load applies configuration values to target. Implementations may partially
	// update target before returning an error. Built-in loaders expect target to
	// be a non-nil pointer supported by their underlying decoder.
	Load(ctx context.Context, target any) error
}

// TypedLoader loads one configuration source into a target of type T.
type TypedLoader[T any] interface {
	// Load applies configuration values to target. Implementations may partially
	// update target before returning an error.
	Load(ctx context.Context, target *T) error
}

type chainLoader struct {
	loaders []Loader
}

type optionalLoader struct {
	loader Loader
}

type typedLoader[T any] struct {
	loader TypedLoader[T]
}

// Chain combines loaders into one ordered, fail-fast Loader. Later loaders see
// and may override changes made by earlier loaders. Chain owns a copy of the
// supplied loader list and reports a nil loader when Load reaches it.
func Chain(loaders ...Loader) Loader {
	return chainLoader{loaders: append([]Loader(nil), loaders...)}
}

// Optional ignores only errors from loader that match fs.ErrNotExist. It
// preserves parse, validation, permission, cancellation, and other errors.
func Optional(loader Loader) Loader {
	return optionalLoader{loader: loader}
}

// Typed adapts a TypedLoader to Loader. Its Load method requires target to have
// type *T and returns an error for a different target type or nil loader.
func Typed[T any](loader TypedLoader[T]) Loader {
	return typedLoader[T]{loader: loader}
}

func (l chainLoader) Load(ctx context.Context, target any) error {
	for i, loader := range l.loaders {
		if isNilValue(loader) {
			return fmt.Errorf("config: loader %d: nil loader", i+1)
		}
		if err := loader.Load(ctx, target); err != nil {
			return fmt.Errorf("config: loader %d: %w", i+1, err)
		}
	}

	return nil
}

func (l optionalLoader) Load(ctx context.Context, target any) error {
	if isNilValue(l.loader) {
		return errors.New("config: optional: nil loader")
	}

	err := l.loader.Load(ctx, target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}

func (l typedLoader[T]) Load(ctx context.Context, target any) error {
	if isNilValue(l.loader) {
		return errors.New("config: typed: nil loader")
	}

	typedTarget, ok := target.(*T)
	if !ok {
		return fmt.Errorf("config: typed loader expects %T, got %T", new(T), target)
	}

	return l.loader.Load(ctx, typedTarget)
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
