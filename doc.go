// Package httpprefix provides small helpers for mounting net/http handlers
// under a configurable URL prefix.
//
// Typical use-case: your service can run at root in local/dev, but under a
// sub-path in production (for example behind a reverse proxy).
//
// The package exposes:
//   - NormalizeRoutePrefix: converts user-configurable values into a canonical
//     prefix ("" or "/prefix")
//   - MountUnderPrefix: mounts handlers under that prefix and applies consistent
//     redirect behavior for the bare prefix
//   - MountUnderPrefixWithOptions: same mount behavior with redirect status
//     code overrides via Option values
package httpprefix
