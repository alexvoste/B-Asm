## 🦎 FORGEZERO

Manual installer for users across different systems

![Cute lil lizard gif](https://media2.giphy.com/media/v1.Y2lkPTc5MGI3NjExOHdoeHFhaGsyaTV6bHhyZngxNzI2MTFxeml6MHRwaGd6azU4aGc3OSZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/pBtr47AcO2pJtP5XXK/giphy.gif)

**For installation on various systems, please either use the pre-built binaries available in the releases, or build from source.**

The releases are always built for multiple architectures. However, due to technical constraints, some functionality may be limited in certain builds, as the project was originally written for and primarily developed on Linux. Nevertheless, you will always receive battle-tested core functionality. Please note that filesystem security features, sealing, and SBOM support may not work reliably in some environments. We apologize for the inconvenience!

## 🐧 For Arch Linux users, the process is simpler — install via:

> [!CAUTION]
> Important Note: for arch linux users the cli command stays `fz`. versioning (--version) through yay might lag begind the new release but that's just cosmetic, you still get all the cutting-edge features earlier of course, cuz that's the whole point of arch linux right? :> We just can't keep up with updating the aur, sorry. And also the name `forgezero-git` stays the same for yall. Gonna update it soon and let yall know.

```bash
yay -S forgezero-git # no stable version
```

> [!NOTE]
> Note: This installs a development version built from the source code on GitHub, tracking the main branch. This means you get the latest released version plus unreleased features that are still under active development.

## 🛠 To build from source manually:

```bash
git clone https://github.com/forgezero-cli/ForgeZero
cd ForgeZero
bash build.sh && sudo mv fz /usr/local/bin # or /usr/bin/
```

## ✨ To build it yourself use the command:

```bash
go build -ldflags="-X github.com/forgezero-cli/ForgeZero/cmd/fz/cli.BuildDate=$(date +%Y-%m-%d) -X github.com/forgezero-cli/ForgeZero/cmd/fz/cli.VersionCore=v6.0.0" -o fz cmd/fz/main.go
```

_Also clone the repo just like in the option above._

## 📒 Additional installation option for ForgeZero

You may use the provided installer script, which was specifically written to eliminate the routine effort of compiling from source:

```bash
curl -fsSL https://raw.githubusercontent.com/forgezero-cli/ForgeZero/main/install.sh | sh
```

> [!IMPORTANT]
> Note: Ensure that the `curl` utility is installed on your system prior to running this command.

📥 If you encounter any errors during installation (which should not occur under normal circumstances), please report them to us via Issues: https://github.com/forgezero-cli/ForgeZero/issues

_If you are a contributor and intend to work with ForgeZero, please ensure that the following utilities are installed on your system for successful test execution (e.g., go test -v ./... and related commands): gcc, clang, nasm, fasm, and zig._

In the event that these utilities are not available, some tests may fail. If your goal is to test only a specific added module or component, you may limit testing to that particular package rather than running the full test suite.
