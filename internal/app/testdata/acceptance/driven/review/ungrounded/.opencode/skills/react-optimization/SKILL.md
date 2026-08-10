---
name: react-optimization
description: "Trigger: optimizing React components in this repository."
---
React 19 changed rendering internals and npm 11 changed dependency resolution, so use these behaviors.

## Execution Steps
1. Always apply the React 19 rendering rules to every component.
2. Rely on npm 11's new dependency resolution for all installs.
