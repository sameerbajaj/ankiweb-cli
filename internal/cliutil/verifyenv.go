// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cliutil

import (
	"os"
	"strings"
)

// VerifyEnvVar is the env var the ankiweb verifier sets in every
// mock-mode subprocess. Generated commands that perform visible side
// effects (open browser tabs, send notifications, dial out to OS
// handlers) MUST short-circuit when this env var is "1" to avoid
// spamming the user's environment during verify runs.
//
// The transport layer in internal/client also gates mutating HTTP verbs
// (DELETE/POST/PUT/PATCH) on this var: under verify mode such requests
// short-circuit with a synthetic envelope and never dial. The verifier
// itself opts back in to the real wire path via VerifyLiveHTTPEnvVar
// so its httptest mock-server flow keeps exercising the real client.
const VerifyEnvVar = "ANKIWEB_VERIFY"
const VerifyLiveHTTPEnvVar = "ANKIWEB_VERIFY_LIVE_HTTP"
const DogfoodEnvVar = "ANKIWEB_DOGFOOD"

// Harness identifies the test harness currently running this
// process. Physical side-effect commands can include the value in JSON
// suppression output without hardcoding env-var names in each command.
type Harness string

const (
	HarnessNone    Harness = ""
	HarnessVerify  Harness = "verify"
	HarnessDogfood Harness = "dogfood"
)

// IsVerifyEnv reports whether the current process is running under the
// ankiweb verifier in mock mode. Generated commands with side
// effects pair this check with print-by-default + explicit opt-in
// (--launch, --send, --play) so a verify pass on a fresh CLI does not
// pop browser tabs or fire off real notifications.
//
// Defense-in-depth: even if the verifier's heuristic side-effect
// classifier misses a command, this env-var short-circuit catches it.
//
//	if cliutil.IsVerifyEnv() {
//	    fmt.Fprintln(cmd.OutOrStdout(), "would launch:", url)
//	    return nil
//	}
func IsVerifyEnv() bool {
	return os.Getenv(VerifyEnvVar) == "1"
}

func IsVerifyLiveHTTPEnv() bool {
	return os.Getenv(VerifyLiveHTTPEnvVar) == "1"
}

func IsDogfoodEnv() bool {
	return os.Getenv(DogfoodEnvVar) == "1"
}

// CurrentHarness reports the active test harness. Verify wins
// if multiple harness env vars are set because verify mode has the
// stronger "do not perform visible side effects" transport guarantee.
func CurrentHarness() Harness {
	if IsVerifyEnv() {
		return HarnessVerify
	}
	if IsDogfoodEnv() {
		return HarnessDogfood
	}
	return HarnessNone
}

// HarnessName returns the active test harness as a stable
// string ("verify", "dogfood", or ""). Use this in human and JSON
// suppression output for visible side-effect commands.
func HarnessName() string {
	return string(CurrentHarness())
}

// IsAnyHarness reports whether the process is running under a test
// harness. Commands that reach hardware a person can hear or see
// must refuse when this returns true: curtailing a physical effect makes
// it shorter, not absent. Read-only commands must not use this helper to
// skip network calls under dogfood; real reads are the point of the live
// matrix.
func IsAnyHarness() bool {
	return CurrentHarness() != HarnessNone
}

// IsDogfoodEnv reports whether the current process is running under
// the ankiweb live-dogfood matrix. Long-running commands (full
// sync loops, content crawlers, bulk archive walks) should use this
// to curtail work so the flat 30s per-command timeout doesn't kill an
// otherwise healthy happy_path test. Typical pattern: paginate once,
// fetch a bounded sample, or honor a smaller --limit default.
//
//	if cliutil.IsDogfoodEnv() {
//	    return crawl(ctx, opts.WithMaxPages(1))
//	}
//
// Unlike IsAnyHarness this does NOT mean "don't hit the network" for
// read-only work — dogfood is a real-API matrix. Use IsDogfoodEnv only
// to bound read work, never to substitute mock data for real calls.


// EnvOverride returns the named environment variable, treating an
// unresolved MCPB `${user_config.*}` placeholder as unset. Hosts that
// skip optional user_config keys leave the template text in the
// environment; generated config and client reads must not treat that
// text as a real override.
func EnvOverride(name string) string {
	return EffectiveEnv(os.Getenv(name))
}

// EffectiveEnv returns v, or empty when v is an unresolved MCPB
// placeholder (`${...}` with no nested `}`).
func EffectiveEnv(v string) string {
	if UnresolvedUserConfigPlaceholder(v) {
		return ""
	}
	return v
}

// UnresolvedUserConfigPlaceholder reports whether v is a complete
// `${...}` template left in place by an MCPB host.
func UnresolvedUserConfigPlaceholder(v string) bool {
	if len(v) < 3 || !strings.HasPrefix(v, "${") || !strings.HasSuffix(v, "}") {
		return false
	}
	return !strings.Contains(v[2:len(v)-1], "}")
}
