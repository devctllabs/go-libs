package telemetry

import (
	"fmt"
	"os"
)

const (
	tracesExporterEnv  = "OTEL_TRACES_EXPORTER"
	metricsExporterEnv = "OTEL_METRICS_EXPORTER"
	generalProtocolEnv = "OTEL_EXPORTER_OTLP_PROTOCOL"
	tracesProtocolEnv  = "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"
	metricsProtocolEnv = "OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"
)

type exporterKind uint8

const (
	exporterNone exporterKind = iota
	exporterOTLP
)

type otlpProtocol uint8

const (
	protocolHTTPProtobuf otlpProtocol = iota
	protocolGRPC
)

type signalConfig struct {
	exporter exporterKind
	protocol otlpProtocol
}

type signalsConfig struct {
	traces  signalConfig
	metrics signalConfig
}

func readSignalConfig() (signalsConfig, error) {
	traces, err := readSignal(tracesExporterEnv, tracesProtocolEnv)
	if err != nil {
		return signalsConfig{}, err
	}
	metrics, err := readSignal(metricsExporterEnv, metricsProtocolEnv)
	if err != nil {
		return signalsConfig{}, err
	}
	return signalsConfig{traces: traces, metrics: metrics}, nil
}

func readSignal(exporterEnv, protocolEnv string) (signalConfig, error) {
	exporter, err := readExporter(exporterEnv)
	if err != nil {
		return signalConfig{}, err
	}
	if exporter == exporterNone {
		return signalConfig{exporter: exporterNone}, nil
	}
	protocol, err := readProtocol(protocolEnv)
	if err != nil {
		return signalConfig{}, err
	}
	return signalConfig{exporter: exporterOTLP, protocol: protocol}, nil
}

func readExporter(key string) (exporterKind, error) {
	switch value := os.Getenv(key); value {
	case "", "otlp":
		return exporterOTLP, nil
	case "none":
		return exporterNone, nil
	default:
		return exporterNone, fmt.Errorf("unsupported %s value %q: expected otlp or none", key, value)
	}
}

func readProtocol(signalKey string) (otlpProtocol, error) {
	value := os.Getenv(signalKey)
	key := signalKey
	if value == "" {
		value = os.Getenv(generalProtocolEnv)
		key = generalProtocolEnv
	}
	if value == "" {
		return protocolHTTPProtobuf, nil
	}
	switch value {
	case "http/protobuf":
		return protocolHTTPProtobuf, nil
	case "grpc":
		return protocolGRPC, nil
	default:
		return protocolHTTPProtobuf, fmt.Errorf("unsupported %s value %q: expected http/protobuf or grpc", key, value)
	}
}
