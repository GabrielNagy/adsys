---
name: Report an issue
about: Create a bug report to fix an existing issue.
title: ''
labels: ''
assignees: ''

---
>**Please do not report security vulnerabilities here**  
>Use [launchpad ADSys private bugs](https://bugs.launchpad.net/ubuntu/+source/adsys/+filebug) which is monitored by our security team. On Ubuntu machines, it’s best to use `ubuntu-bug adsys` to collect relevant information.

**Thank you in advance for helping us to improve ADSys!**  
Please read through the template below and answer all relevant questions. Your additional work here is greatly appreciated and will help us respond as quickly as possible. For general support or usage questions, use [Ubuntu Discourse](https://discourse.ubuntu.com/c/desktop/8). Finally, to avoid duplicates, please search existing Issues before submitting one here.

By submitting an Issue to this repository, you agree to the terms within the [Ubuntu Code of Conduct](https://ubuntu.com/community/code-of-conduct).

## Description

> Provide a clear and concise description of the issue, including what you expected to happen.

## Reproduction

> Detail the steps taken to reproduce this error, what was expected, and whether this issue can be reproduced consistently or if it is intermittent.
>
> Where applicable, please include:
>
> * Code sample to reproduce the issue
> * Log files (redact/remove sensitive information)
> * Application settings (redact/remove sensitive information)
> * Screenshots

### Environment

> Please provide the following:

#### For Ubuntu users, please run and copy the following

1. `ubuntu-bug adsys --save=/tmp/report`
1. Copy paste below `/tmp/report` content:

```raw
COPY REPORT CONTENT HERE.
```

#### Relevant AD information

If AD authentication works but adsys fails to fetch GPOs (e.g. you see `can't get policies` errors on login), please perform the following steps:

1. Add the following to `/etc/samba/smb.conf`:
```
log level = 10
```
2. Run `sudo login user@domain` in a terminal, replacing with your AD credentials
3. Paste the output in the bug report

#### Installed versions

* OS: (`/etc/os-release`)
* ADSys version: (`adsysctl version` output)

#### Additional context

> Add any other context about the problem here.
