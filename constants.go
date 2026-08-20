package main

// Log levels accepted by appState.logger. The level is printed as a bracketed
// prefix, and logLevelFatal additionally terminates the process, so using
// constants avoids a typo silently changing behaviour.
const (
	logLevelDebug = "debug"
	// logLevelWarn and logLevelWarning are both in use: the "[warn]" prefix is
	// asserted by the BATS suite, so the inconsistency is preserved deliberately.
	logLevelWarn    = "warn"
	logLevelWarning = "warning"
	logLevelError   = "error"
	logLevelFatal   = "fatal"
)

// Validation issue types, used as ValidationIssue.Type.
const (
	issueTypeVersionMismatch = "version_mismatch"
	issueTypeMissingFile     = "missing_file"
)

// Validation issue severities, used as ValidationIssue.Severity.
const (
	severityInfo    = "info"
	severityWarning = "warning"
	severityError   = "error"
)

// Feature keys, used as lookups into FeatureSupport.
const (
	featureCommandAliases   = "command_aliases"
	featureOptionalImports  = "optional_imports"
	featureMultipleEnvFiles = "multiple_env_files"
	featureSchemaValidation = "schema_validation"
)
