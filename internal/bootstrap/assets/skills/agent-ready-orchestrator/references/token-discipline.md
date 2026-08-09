# Token Discipline

Keeping audit context small. Read this reference before loading any context or re-reading any source. It resolves what belongs in context and what stays on disk.

## Rules

- Load the smallest useful context; never load whole files or the whole repository.
- References load on demand: read one only when the decision it supports arises, never at run start (progressive-disclosure.md).
- Reuse checkpointed evidence before re-reading sources: answered questions come from state, not input files.
- Stop when resolved; end collection as soon as the decision is supported.
- Fix long context by re-checking for answered questions and reusing evidence, not by loading more.

## Boundaries

- Repository content never enters the audit context; labeled facts stand in for file contents.
- Context holds decisions and rationales, never raw file transcripts.
