---
title: Middleware
weight: 60
---
To accelerate your development, the Flamego core team and the community have built some useful middleware in addition to the [core services](/core-services) that are builtin to the core framework.

- [template](/middleware/template) for rendering HTML using Go template.
- [session](/middleware/session) for managing user sessions.
- [recaptcha](/middleware/recaptcha) for providing [Google reCAPTCHA](https://www.google.com/recaptcha/about/) verification.
- [csrf](/middleware/csrf) for generating and validating CSRF tokens.
- [cors](/middleware/cors) for configuring [Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).
- [binding](/middleware/binding) for request data binding and validation.
- [gzip](/middleware/gzip) for Gzip compression to responses.
- [cache](/middleware/cache) for managing cache data.
- [brotli](/middleware/brotli) for Brotli compression to responses.
- [auth](/middleware/auth) for providing basic and bearer authentications.
- [i18n](/middleware/i18n) for providing internationalization and localization.
- [captcha](/middleware/captcha) for generating and validating captcha images.
- [hcaptcha](/middleware/hcaptcha) for providing [hCaptcha](https://www.hcaptcha.com/) verification.

{{< callout type="info" >}}
If you notice any middleware that is missing from the list, please don't hesitate to [send a pull request to this page](https://github.com/flamego/flamego.dev/edit/main/docs/middleware/README.md)!
{{< /callout >}}