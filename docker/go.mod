module dagger/docker

go 1.24.0

require (
	emperror.dev/errors v0.8.1
	github.com/coreos/go-semver v0.3.1
	github.com/creasty/defaults v1.8.0
	github.com/gookit/validate v1.5.4
)

require github.com/stretchr/testify v1.11.1 // indirect

require (
	github.com/gookit/filter v1.2.2 // indirect
	github.com/gookit/goutil v0.6.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc => go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.14.0

replace go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp => go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.14.0

replace go.opentelemetry.io/otel/log => go.opentelemetry.io/otel/log v0.14.0

replace go.opentelemetry.io/otel/sdk/log => go.opentelemetry.io/otel/sdk/log v0.14.0
