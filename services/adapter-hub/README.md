# adapter-hub

Integration adapter hub for external partners. Supports demo and live modes for NID and MFI providers.

**Port**: 8013  
**Mode**: controlled by `MESOB_ADAPTER_MODE` (DEMO | LIVE)  
**Auth**: INTERNAL mTLS for /adapters/nid and /adapters/mfi; PARTNER SIGNED for webhooks

## Local run (demo mode)
`MESOB_ADAPTER_MODE=DEMO go run ./cmd/server`
