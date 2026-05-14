---
title: Middleware
weight: 60
---
To accelerate your development, the Flamego core team and the community have built some useful middleware in addition to the [core services](../core-services) that are builtin to the core framework.

- [template](template) for rendering HTML using Go template.
- [session](session) for managing user sessions.
- [recaptcha](recaptcha) for providing [Google reCAPTCHA](https://www.google.com/recaptcha/about/) verification.
- [csrf](csrf) for generating and validating CSRF tokens.
- [cors](cors) for configuring [Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).
- [binding](binding) for request data binding and validation.
- [gzip](gzip) for Gzip compression to responses.
- [cache](cache) for managing cache data.
- [brotli](brotli) for Brotli compression to responses.
- [auth](auth) for providing basic and bearer authentications.
- [i18n](i18n) for providing internationalization and localization.
- [captcha](captcha) for generating and validating captcha images.
- [hcaptcha](hcaptcha) for providing [hCaptcha](https://www.hcaptcha.com/) verification.

{{< callout type="info" >}}
If you notice any middleware that is missing from the list, please don't hesitate to [send a pull request to this page](https://github.com/flamego/flamego.dev/edit/main/docs/middleware/README.md)!
{{< /callout >}}