package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

func TestBuildResourceAppliesPrecedence(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=env-name,service.version=env-version,deployment.environment.name=env-environment,service.instance.id=env-instance,custom.key=custom-value")

	res, err := buildResource(context.Background(), validConfig())
	require.NoError(t, err)
	require.Equal(t, "orders", resourceString(t, res.Set(), semconv.ServiceNameKey))
	require.Equal(t, "1.2.3", resourceString(t, res.Set(), semconv.ServiceVersionKey))
	require.Equal(t, "test", resourceString(t, res.Set(), semconv.DeploymentEnvironmentNameKey))
	require.Equal(t, "env-instance", resourceString(t, res.Set(), semconv.ServiceInstanceIDKey))
	require.Equal(t, "custom-value", resourceString(t, res.Set(), attribute.Key("custom.key")))
}

func TestBuildResourceGeneratesServiceInstanceID(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")

	res, err := buildResource(context.Background(), validConfig())
	require.NoError(t, err)
	require.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, resourceString(t, res.Set(), semconv.ServiceInstanceIDKey))
}

func resourceString(t *testing.T, set *attribute.Set, key attribute.Key) string {
	t.Helper()
	value, ok := set.Value(key)
	require.True(t, ok)
	return value.AsString()
}
