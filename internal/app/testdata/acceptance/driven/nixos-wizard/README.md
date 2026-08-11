# nixos-wizard

A NixOS configuration wizard with a Ratatui terminal interface.

## Layout

- `Cargo.toml` / `Cargo.lock` — Rust workspace with `ratatui` 0.29.
- `flake.nix` / `flake.lock` — Nix flake for the dev shell.
- `src/main.rs` — terminal UI entry point.

## Development

Use the Nix dev shell: `nix develop`, then `cargo run`.

## Frontend

The web dashboard uses pnpm:

```sh
pnpm install
pnpm run dev
```
