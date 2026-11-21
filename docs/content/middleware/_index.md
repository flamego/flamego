---
title: Middleware
weight: 60
---
To accelerate your development, the Flamego core team and the community have built some useful middleware in addition to the [core services](/core-services) that are builtin to the core framework.

- [template](/middleware/template) for rendering HTML using Go template.
- [session](/middleware/session) for managing user sessions.
- [cache](/middleware/cache) for managing cache data.
- [binding](/middleware/binding) for request data binding and validation.
- [csrf](/middleware/csrf) for generating and validating CSRF tokens.
- [cors](/middleware/cors) for configuring [Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).
- [gzip](/middleware/gzip) for Gzip compression to responses.
- [brotli](/middleware/brotli) for Brotli compression to responses.
- [i18n](/middleware/i18n) for providing internationalization and localization.
- [auth](/middleware/auth) for providing basic and bearer authentications.
- [altcha](/middleware/altcha) for providing self-hosted [ALTCHA](https://altcha.org) proof-of-work captchas.
- [recaptcha](/middleware/recaptcha) for providing [Google reCAPTCHA](https://www.google.com/recaptcha/about/) verification.
- [hcaptcha](/middleware/hcaptcha) for providing [hCaptcha](https://www.hcaptcha.com/) verification.
- [captcha](/middleware/captcha) for generating and validating captcha images.
- [sse](/middleware/sse) for sending server events to web clients.

{{< callout type="info" >}}
If you notice any middleware that is missing from the list, please don't hesitate to [send a pull request to this page](https://github.com/flamego/flamego/edit/main/docs/content/middleware/_index.md)!
{{< /callout >}}
