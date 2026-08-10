# Driven Review Cohorts

Unseeded repository content for the driven review proof (`TestDrivenReview` in
`internal/app/driven_review_test.go`). The harness copies each cohort into a
fresh temporary Git repository, installs the harness with `agent-ready init`,
and runs the real `opencode run --command agent-ready review`. Each cohort
seeds one candidate skill under `.opencode/skills/` as the artifact under
review; no conclusions, expected verdicts, or exemptions are seeded. The
structural oracle observes only events, state, and artifacts the review
itself produces.

- `grounded/` — a Go module plus a candidate skill whose version-sensitive
  toolchain knowledge cites current versioned evidence (OpenCode 1.18.15
  official documentation), so the review MUST accept it through the External
  Verification Gate.
- `ungrounded/` — a Go module plus a candidate skill that embeds framework
  claims without versioned evidence or exemption, so the review MUST reject
  it (gate failure) or explicitly handle the exemption.
