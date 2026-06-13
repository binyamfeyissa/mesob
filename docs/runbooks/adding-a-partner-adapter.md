# Runbook: Adding a Partner Adapter

## Overview
All external partner integrations go through the `adapter-hub` service. New partners add a new provider implementation without changing domain services.

## Steps

### 1. Add provider interface implementation
In `services/adapter-hub/internal/infra/providers/`, create `{partner}_provider.go`:

```go
package providers

type AcmeNIDProvider struct {
    endpoint string
    apiKey   string
}

func (p *AcmeNIDProvider) Verify(ctx context.Context, fan, name, dob string) (bool, float64, error) {
    // TODO: call Acme API
}
```

### 2. Add config variables
In `internal/infra/config/config.go`, add:
```go
AcmeEndpoint string
AcmeAPIKey   string
```
And corresponding env vars: `MESOB_ADAPTER_ACME_ENDPOINT`, `MESOB_ADAPTER_ACME_API_KEY`.

### 3. Wire in main.go
Switch on `cfg.Mode` or `cfg.NIDProvider`:
```go
var nidProvider app.NIDProvider
switch cfg.NIDProvider {
case "acme":
    nidProvider = providers.NewAcmeNIDProvider(cfg.AcmeEndpoint, cfg.AcmeAPIKey)
default:
    nidProvider = providers.NewDemoNIDProvider()
}
```

### 4. Add circuit breaker
Wrap provider calls in a circuit breaker. On open circuit, return `ErrProviderUnavailable` (retryable).

### 5. Update adapter status endpoint
The `GET /adapters/status` endpoint should reflect the new adapter's mode and breaker state.

### 6. Test with DEMO mode first
Set `MESOB_ADAPTER_MODE=DEMO` and verify the full flow before switching to LIVE.

### 7. Add partner webhook handler
If the partner sends callbacks, add a handler in `partnerWebhook` for `r.PathValue("partner") == "acme"`. Verify HMAC signature before processing.
