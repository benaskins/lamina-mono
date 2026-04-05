# Fix axon-auth availability

**Area:** axon-auth, aurelia
**Discovered:** 2026-04-03

The auth service at `auth.hestia.internal` is returning "auth service
unavailable" from synd's perspective. This blocks all API operations
that require authentication (post, approve, delete, rebuild).

Investigate whether the auth service is running, whether the TLS chain
is valid, and whether synd can reach it. Check aurelia status for auth.
