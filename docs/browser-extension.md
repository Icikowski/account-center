# Browser extension

The [`account-center-browser-extension`](https://git.sr.ht/~icikowski/account-center-browser-extension) repository contains a Manifest V3 extension for **Firefox** and **Chrome** (including Chromium-based browsers). It shows the signed-in user and a filterable list of the services available to them, using the [API](api.md).

It deliberately does little: no page access, no content scripts, no background worker, no tokens.

## Stores

- Firefox: [Firefox Add-ons](https://addons.mozilla.org/firefox/addon/account-center/)
- Chrome (including Chromium-based browsers): [Chrome Web Store](https://chromewebstore.google.com/detail/ngjhhagcpndonjabkbdmgfgephejmkgi/)

## How it works

1. The user enters the instance address in the extension settings.
2. The browser asks to grant the extension access to **that host only**.
3. The popup calls `/api/v1/session` and `/api/v1/services` with the instance's own session cookie.
4. When signed out, the popup opens `/login` in a regular tab, so sign-in happens through the normal OIDC flow.

Nothing is stored by the extension except the instance address.

## Requirements on the server

- The instance must support the [API](api.md) (version `1.10.0` or later).
- The instance must be served over HTTPS (HTTP is accepted for `localhost` and `127.0.0.1` only).
- The browser must send the session cookie (`SameSite=Lax`) on extension requests. This works for extensions that hold host access to the instance's origin. Browsers configured to block third-party cookies globally may need an exception for the instance.
- No additional environment variables are needed.

## Limitations

- One instance per browser profile.
- Status is checked when the popup opens; there is no toolbar badge or background polling.

## See also

- [API](api.md)
- [Configuration overview](configuration.md)
- [Services catalog](services-catalog.md)
