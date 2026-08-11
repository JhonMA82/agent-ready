{
  "description": "NixOS wizard flake",
  "inputs": {
    "nixpkgs": {
      "type": "github",
      "owner": "NixOS",
      "repo": "nixpkgs"
    }
  },
  "outputs": {
    "devShells": {
      "default": {
        "packages": ["rustc", "cargo"]
      }
    }
  }
}
