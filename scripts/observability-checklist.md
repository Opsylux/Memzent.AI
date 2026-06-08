# Production Observability Checklist (Sprint 5)

Manual sign-off before go-live. Pair with [GO_LIVE_CHECKLIST.md](../GO_LIVE_CHECKLIST.md) §5.5.

## Gateway metrics

- [ ] Prometheus (or equivalent) scrapes `GET /metrics` on the gateway service
- [ ] Dashboard panels exist for: request rate, cache hit ratio, p95 latency, error rate

## Health probes

- [ ] Liveness: `GET /healthz`
- [ ] Readiness: `GET /readyz` (Valkey + Postgres + Router)
- [ ] Alert when `readyz` fails for > 2 minutes

## Audit logs

- [ ] Confirm `main.go` audit retention job runs (30-day purge)
- [ ] Log shipping configured (CloudWatch, Datadog, Loki, etc.)

## Billing & webhooks

- [ ] Alert on Stripe webhook handler errors (`/v1/billing/webhook`)
- [ ] Alert on sustained `payment required` rate spikes

## Vector store DR

- [ ] `SNAPSHOT_INTERVAL_HOURS=24` set in production router env
- [ ] Qdrant snapshot destination configured (S3 or verified local volume backups)

## Cache anomalies

- [ ] Alert when org-level cache hit ratio drops > 30% below 7-day baseline (possible Valkey outage or key eviction)
