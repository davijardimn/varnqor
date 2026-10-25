# Security Policy

Varnqor is currently pre-alpha.

## Reporting a vulnerability

Please do not publish sensitive vulnerability details in a public issue.

For now, contact the repository owner through GitHub before disclosing details publicly. A dedicated private reporting channel will be added before the first public release.

## Scope

Security-sensitive areas include:

- parsing untrusted local agent logs;
- filesystem traversal and path handling;
- desktop bridge methods exposed to the frontend;
- future provider credentials or API tokens;
- any future update mechanism.

Varnqor should never be used to extract credentials, bypass authentication, or circumvent provider usage limits.
