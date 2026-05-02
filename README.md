# agent-michi

`agent-michi` is a specialized orchestration agent designed exclusively for the **Nukemichi** private network management ecosystem. It handles remote host inspection and service lifecycle management on the server side.

## Ecosystem Role

This repository is the backend component of the [Nukemichi](https://github.com/Nukemichi/nukemichi-android) project.

- **For Android contributors:** You generally don't need to build this manually. Use the `scripts/ensure-agent-binaries.sh` script in the main repository to sync pre-built and compressed binaries.
- **For Agent contributors:** Use the provided `Makefile` to build and test your changes across all supported architectures.

## Development & Build Manual

### Prerequisites
- **Go 1.22+**
- **Make**
- **zstd** (for production-ready compression)

### Using the Makefile
The project includes a `Makefile` to simplify cross-compilation with optimized flags (`-s -w` to minimize binary size).

- **Build all targets:**
  ```bash
  make all
  ```
- **Build for specific architectures:**
  ```bash
  make linux-arm64    # Modern ARM servers
  make linux-amd64    # Standard x86 servers
  make linux-armv7    # Legacy ARM hardware
  make darwin-arm64   # Local testing on Apple Silicon
  ```

Build artifacts will be placed in the `build/` directory with the naming convention: `agent-michi-{os}-{arch}`.

### Compression for Nukemichi Bundling
The Android application expects the binaries to be compressed using `zstd` before being placed in the assets. If you are manually updating the app's bundled binaries, compress them as follows:
```bash
zstd --ultra -22 build/agent-michi-linux-arm64 -o build/agent-michi.zst
```

## Integration Details
Once built and compressed, the `.zst` files are typically synced into the Android project using the automation scripts found in the `nukemichi-android` repository. They end up in:
`core/agent/impl/src/main/assets/bin/{arch}/agent-michi.zst`

## License
See [LICENSE](LICENSE).
