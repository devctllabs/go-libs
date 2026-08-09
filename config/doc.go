// Package config loads configuration into caller-owned values from ordered,
// composable sources.
//
// # Loading and precedence
//
// Pass a non-nil pointer to the application config value to Loader.Load. Chain
// applies loaders in the supplied order, so callers express increasing
// precedence directly: defaults, files, dotenv data, process environment, then
// explicit runtime overrides. Loading is fail-fast, and the target may be
// partially updated when a loader returns an error; load into a temporary value
// and validate it before publishing it to the application.
//
// Defaults reads "default" tags. YAML and TOML use their corresponding tags and
// overlay only fields present in the input; maps and slices are replaced as
// whole values. DotEnv, EnvMap, and OSEnv decode "env" tags. DotEnv does not
// modify the process environment, EnvMap owns a snapshot of the supplied map,
// and OSEnv reads the process environment when Load is called.
//
// # Inputs and custom sources
//
// Path and FromFS provide inputs backed by an operating system path or fs.FS.
// Optional suppresses only errors that match fs.ErrNotExist, which makes it
// suitable for optional config files without hiding parse or permission errors.
//
// Implement Input when an existing format must read from another location.
// Implement Loader for a new format or source, or implement TypedLoader and use
// Typed when the source is specific to one application config type. An S3,
// Vault, CLI, or other source is another loader placed at its chosen precedence;
// it does not require changes to Chain.
//
// CLI adapters should represent presence explicitly, commonly with pointer
// fields, so an absent option remains distinguishable from explicit false, zero,
// an empty string, or a nil-capable value. The application composition root
// should load and validate the final value once before constructing runtime
// resources.
//
// Generated gomock implementations of Input and Loader are available in the
// config/mocks subpackage for direct consumers of those interfaces.
package config
