
## Mock Clarifying Questions & Design Rationale

### Architecture & Distribution Questions

**Q: How should we distribute updates - through a centralized server or decentralized approach?**
*Answer: Centralized via GitHub releases with a JSON manifest. This leverages existing infrastructure, provides reliable CDN distribution, and simplifies the update discovery mechanism.*

**Q: Should updates be automatic or require user confirmation?**
*Answer: Require user confirmation with a Y/N prompt. Respect user agency, prevents unexpected disruptions during critical work, and allows users to defer updates if needed.*

**Q: How should we handle different release channels (stable, beta, alpha)?**
*Answer: Implement strict channel isolation with environment variable opt-ins. Users stay in their chosen channel unless they explicitly enable cross-channel updates via `CALC_ALLOW_ALPHA` or `CALC_ALLOW_BETA` flags.*

### Version Management Questions

**Q: What versioning scheme should we use?**
*Answer: Semantic versioning (vX.Y.Z) with pre-release suffixes (-alpha, -beta). This provides clear upgrade paths and allows proper version comparison logic.*

**Q: How should we handle version comparison with pre-releases?**
*Answer: Custom semantic version parser that treats stable > beta > alpha, with proper numerical comparison of major.minor.patch components. This prevents users from accidentally downgrading.*

**Q: Should we maintain a version history or just track the latest?**
*Answer: Maintain a manifest of the last 50 releases. This provides rollback options, supports different release channels, and gives users visibility into update history.*

### Security & Safety Questions

**Q: How do we ensure downloaded updates are legitimate and haven't been tampered with?**
*Answer: Validate executable magic bytes (ELF, Mach-O, PE) and test the binary with `--version` before installation. While not cryptographically secure, this catches basic corruption. Safer future implementations should have the build pipeline create a SHA1 hash and have the client verify this before installation.*

**Q: What happens if an update fails or corrupts the binary?**
*Answer: Create a backup of the current version before updating (named with version suffix). If download/validation fails, cleanup temp files and preserve the working binary.*

**Q: Should we support rollback to previous versions?**
*Answer: Implicit rollback via backups. Each update creates a `.bak` file, allowing manual recovery. Full rollback would require more complex state management and potentially more then 50 builds saved in the manifest system for long-term production support.*

### Platform Compatibility Questions

**Q: How do we handle cross-platform binary distribution?**
*Answer: Build separate binaries for each platform (linux/amd64, windows/amd64, darwin/amd64, darwin/arm64) and use runtime.GOOS to select the appropriate download URL.*

**Q: Should the updater work offline or require internet connectivity?**
*Answer: Require internet for update checks but fail gracefully with warnings. The core calculator functionality should work offline.*

### Testing Strategy Questions

**Q: How extensively should we test the update mechanism?**
*Answer: Comprehensive unit tests with edge cases, mock HTTP responses, and channel isolation verification. The update logic is complex enough to warrant thorough testing, especially version comparison and channel switching rules.*

**Q: What about integration testing for the actual download/install process?**
*Answer: Mock the HTTP layer and file operations for unit tests. Real integration testing would require actual releases, which is complex for CI. Focus on unit testing the logic and manual testing of real releases.*

### CI/CD & Build Questions

**Q: How should CI/CD handle versioning during builds?**
*Answer: Extract version from git tags during release builds, with fallback to git describe for development builds. Pass explicit VERSION environment variable to override Make's git-based version detection.*

**Q: Should we auto-update the version manifest when releasing?**
*Answer: Yes, but with conflict resolution. Use a separate CI job that rebuilds the entire manifest from GitHub API data, with exponential backoff retry logic to handle concurrent releases.*

**Q: How do we handle the chicken-and-egg problem of releasing a first version?**
*Answer: The updater gracefully handles missing manifests or API failures. Initial deployment doesn't depend on the update mechanism - it's additive functionality.*

### User Experience Questions

**Q: What should happen during the first run before any updates exist?**
*Answer: Show a warning that no manifest was found, but continue with normal calculator operation. Updates are supplementary to core functionality.*

**Q: How verbose should update checks be?**
*Answer: Moderate verbosity - show "Checking for updates..." progress, version information, and clear prompts. Avoid spamming users but provide enough information for transparency.*

**Q: Should we support silent/headless updates for automation?**
*Answer: Not implemented, but the architecture supports it. Environment variables could control update behavior, and the prompt logic could be bypassed in automation scenarios.*

### Error Handling & Reliability Questions

**Q: What if GitHub is down or releases are unavailable?**
*Answer: Graceful degradation with warnings. The calculator continues to function normally, and update checks simply fail with informative messages rather than crashing.*

**Q: How do we handle partial downloads or network interruptions?**
*Answer: Download to temporary files first, validate size and content, then atomically replace the binary. Failed downloads are cleaned up automatically.*

**Q: Should we retry failed update attempts?**
*Answer: Not for binary downloads (user can retry manually), but yes for manifest updates in CI (exponential backoff). This balances reliability with avoiding infinite retry loops.*
