# Evidence Labels

Classification discipline for audit findings. Read this reference when labeling a finding or when a decision needs evidence. It resolves how FACT, INFERENCE, and UNKNOWN differ and what each permits.

## Labels (normative)

| Label | Meaning | Permits |
|---|---|---|
| FACT | Directly observed from a deterministic source: file content, helper output, recorded state | Evidence for decisions, including CREATE |
| INFERENCE | Derived from facts with a stated reasoning step | Hypothesis only; never CREATE alone |
| UNKNOWN | Not observed and not derivable | No decision; gather more or ASK_USER |

## Rules

- Label by the weakest permitted claim: an unverified observation is UNKNOWN, not INFERENCE.
- Every INFERENCE names the facts it derives from and the reasoning step.
- Every UNKNOWN names the evidence that would resolve it.
- A fact without a source is not evidence; every FACT names its source.
- Counts such as "N skills generated" are never evidence of quality or necessity.
