# Security policy

## Reporting a vulnerability

If you think you've found a security problem in Ahoy, please report it privately rather than opening a public issue.

Use GitHub's private reporting form: [report a vulnerability](https://github.com/ahoy-cli/ahoy/security/advisories/new). Only the maintainers can see what you send.

It helps if your report includes:

- the Ahoy version (`ahoy --version`) and your operating system
- the steps or a minimal `.ahoy.yml` that shows the problem
- what an attacker could do with it

Ahoy is maintained by volunteers, so replies are best effort. We'll confirm we've received your report, keep you updated while we work on a fix, and credit you in the advisory unless you'd rather stay anonymous. Please give us a reasonable chance to release a fix before you share details publicly.

## Supported versions

Security fixes go into the latest v3 release. Upgrade to it before reporting, in case the problem is already fixed.

## Verifying downloads

Every release ships SHA-256 checksums, and releases from v3.0.2 onwards are signed with cosign. See [Verifying your download](https://ahoy-cli.github.io/guides/getting-started/#verifying-your-download).
