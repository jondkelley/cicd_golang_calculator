
## version.json Structure

version.json stores an array of releases and github workflows release.yml automatically commits new tagged releases to github (storing the last 50 releases)

```json
{
  "releases": [
    {
      "version": "1.0.3",
      "urls": {
        "linux": "https://github.com/user/repo/releases/download/v1.0.3/calc-linux",
        "windows": "https://github.com/user/repo/releases/download/v1.0.3/calc-windows.exe",
        "darwin": "https://github.com/user/repo/releases/download/v1.0.3/calc-macos"
      },
      "isAlpha": false,
      "releaseDate": "2025-05-30T10:30:00Z"
    },
    {
      "version": "1.0.3-alpha",
      "urls": {
        "linux": "https://github.com/user/repo/releases/download/v1.0.3-alpha/calc-linux",
        "windows": "https://github.com/user/repo/releases/download/v1.0.3-alpha/calc-windows.exe",
        "darwin": "https://github.com/user/repo/releases/download/v1.0.3-alpha/calc-macos"
      },
      "isAlpha": true,
      "releaseDate": "2025-05-29T15:20:00Z"
    }
  ]
}
```

## Usage Examples

### Normal User (only gets stable releases):
```bash
./calc
# Can only see stable releases like v1.0.3, v1.0.2, etc.
# Skips v1.0.3-alpha entirely
```

### Developer/Tester (gets alpha releases):
```bash
CALC_ALLOW_ALPHA=1 ./calc
# Can see ALL releases including alphas
# Gets v1.0.3-alpha if it's newer than the current version
```

### Testing Different Tags:
```bash
# Stable release
git tag v1.0.1 && git push origin v1.0.1

# Alpha release  
git tag v1.0.2-alpha && git push origin v1.0.2-alpha
```
