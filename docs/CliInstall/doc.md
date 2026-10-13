# CliInstall

`convexa` is a single static binary: no runtime, no dependencies. Pick your platform,
paste the block, done. Go 1.25+ is needed only to build it from source.

**macOS (Apple Silicon)**

```bash
curl -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/macarm64.bin -o convexa && chmod +x convexa && sudo mv convexa /usr/local/bin/ && convexa version
```

**macOS (Intel)**

```bash
curl -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/mac86.bin -o convexa && chmod +x convexa && sudo mv convexa /usr/local/bin/ && convexa version
```

**Linux (amd64)**

```bash
curl -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/linux86.out -o convexa && chmod +x convexa && sudo mv convexa /usr/local/bin/ && convexa version
```

**Linux (arm64)**

```bash
curl -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/linuxarm64.out -o convexa && chmod +x convexa && sudo mv convexa /usr/local/bin/ && convexa version
```

**Linux (32-bit)**

```bash
curl -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/linuxi32.out -o convexa && chmod +x convexa && sudo mv convexa /usr/local/bin/ && convexa version
```

**Windows (64-bit)** — PowerShell:

```powershell
$dir="$HOME\.local\bin"; New-Item -ItemType Directory -Force -Path $dir | Out-Null; curl.exe -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/windows86.exe -o "$dir\convexa.exe"; [Environment]::SetEnvironmentVariable('PATH', [Environment]::GetEnvironmentVariable('PATH','User') + ";$dir", 'User')
```

**Windows (32-bit)** — PowerShell:

```powershell
$dir="$HOME\.local\bin"; New-Item -ItemType Directory -Force -Path $dir | Out-Null; curl.exe -sL https://github.com/MateusMoutinhoOrg/Convexa/releases/latest/download/windowsi32.exe -o "$dir\convexa.exe"; [Environment]::SetEnvironmentVariable('PATH', [Environment]::GetEnvironmentVariable('PATH','User') + ";$dir", 'User')
```

**From a checkout** — needs Go 1.25+:

```bash
go build -o convexa ./cmd/main && sudo mv convexa /usr/local/bin/ && convexa version
```

The released binaries are the ones `agnos compile --target all` builds and `agnos publish`
uploads. `convexa version` prints the `version` of `AgnosConfig/project.yaml`,
`convexa help` every command — each one is listed in [Commands](../Commands/doc.md).
