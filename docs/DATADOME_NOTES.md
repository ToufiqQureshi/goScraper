# Datadome / anti-bot notes

This library is intentionally small and does not currently include stealth, anti-detection, anti-bot bypass, or browser fingerprint evasion features.

## What this means

- The library can automate a visible Chromium browser and extract page data.
- It is not a guaranteed solution for websites protected by Datadome, Akamai Bot Manager, Kasada, or similar systems.
- If a target site blocks automation, the library may need a different strategy at the application level.

## What we can do in this project

1. Keep the public API simple and predictable.
2. Add better page lifecycle, selector, click, type, wait, and evaluation helpers.
3. Add optional debugging hooks and diagnostic errors.
4. Support retries, timeouts, and structured error reporting.
5. Document when a site is likely to require browser-automation alternatives or human-in-the-loop handling.

## What we should not promise

- "Undetectable browsing"
- "Bypass Datadome"
- "Automatic unlock for all bot-protected sites"

## Recommended approach for real-world Datadome cases

- Use this library only for ordinary browser automation tasks.
- If a site uses strong bot protection, the practical options are usually:
  - use an approved API from the target site,
  - solve the challenge in a browser with manual or semi-automated assistance,
  - use a dedicated browser automation setup with proper operational controls,
  - or redesign the integration to avoid scraping protected content.

## Suggested work for this repo

- Keep feature additions focused on reliability and ergonomics.
- Add tests for API validation and lifecycle behavior.
- Add docs that clearly state the limits of browser automation against anti-bot systems.
- Treat Datadome as a separate problem, not as a library feature that can be casually added.
