---
diataxis: reference
tags:
  - business
  - enterprise
  - operations
  - security
---
# Enterprise readiness evidence

Use this map before relying on Enterprise terms for custom operational, legal, security, or deployment commitments.

!!! warning "Discovery required"
    Enterprise pricing starts at €60,000/year, but the price alone is not a high-availability, support, security, or
    managed-operations promise. Validate required HA, support, security, and operational commitments during discovery
    and capture the agreed scope in the Order Form.

## When Enterprise is appropriate

Enterprise is appropriate only when custom scope is required, such as:

- OEM, resale, or third-party hosting rights
- Regulated deployments or custom legal/security terms
- Multi-region high availability or custom disaster-recovery targets
- Custom SLA, support channel, service credits, or escalation commitments
- Custom deployment/support obligations that exceed Starter or Growth defaults

Use Starter or Growth when published plan scope and default support terms are enough.

## Evidence map

| Area | Question to validate | Evidence |
| --- | --- | --- |
| Kubernetes deployment | Can the buyer deploy and configure the stack in their cluster? | [Deploy with Helm](../how-to/deploy-helm.md) |
| Production readiness | Are launch blockers, rollback, observability, and hardening tracked? | [Production readiness checklist](../operations/production-readiness-checklist.md) |
| Security/procurement | What security artifacts are published, and what requires custom review? | [Security pack](../security/security-pack.md) |
| Support/SLA | What response commitments exist by default or by Order Form? | [SLA and support schedule](../security/sla-schedule.md) |
| Runbooks | How are common operational failures handled? | [Operations runbooks](../operations/index.md) |
| Limitations | Which boundaries and non-goals must be accepted or negotiated? | [Known limitations](../start-here/known-limitations.md) |
| Evaluation evidence | Does the pilot produce an auditable decision trail? | [Evidence](evidence.md) |

## Enterprise discovery checklist

- [ ] Deployment topology and ownership are known, including single-region vs multi-region expectations.
- [ ] HA, backup, disaster recovery, and rollback expectations are written down.
- [ ] Support hours, response targets, escalation path, and service-credit expectations are defined.
- [ ] Security, privacy, data residency, and procurement constraints are listed.
- [ ] Required legal/security terms are identified before Order Form drafting.
- [ ] Known limitations have been reviewed and accepted or captured as custom scope.
- [ ] Pilot/proof-kit evidence has been reviewed before production commitments are made.

## Order Form implications

Enterprise commitments should be explicit in the Order Form or attached schedules. In particular, do not assume that
Enterprise includes multi-region HA, 24x7 support, service credits, custom legal riders, or managed operations unless
those commitments are stated in writing.

## Read next

- Pricing: [Pricing](../pricing/index.md)
- Order form template: [Order form template](../licensing/order_form.md)
- Procurement checklist: [Procurement checklist](procurement-checklist.md)
- Security pack: [Security pack](../security/security-pack.md)
