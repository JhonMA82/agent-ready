# Driven Audit Cohort

Unseeded repository content for the driven audit proof
(`TestDrivenAudit` in `internal/app/driven_audit_test.go`). The harness copies
these files into a fresh temporary Git repository, installs the harness with
`agent-ready init`, and runs the real `opencode run --command agent-ready
audit`. No conclusions, expected outputs, or verdicts are seeded anywhere in
this fixture; the structural oracle observes only events, state, and artifacts
the audit itself produces.
