# Flamego documentation

The site is built with [Hugo](https://gohugo.io/) and the [Hextra](https://imfing.github.io/hextra/) theme.

## Prerequisites

- **Hugo extended** ≥ 0.160 (`brew install hugo`)
- **Go** ≥ 1.20 (Hextra is pulled in as a Hugo module)

## Local preview

From the `docs/` directory:

```sh
hugo server
```

Then open <http://localhost:1313/>.

To preview on other devices on your network:

```sh
hugo server --bind 0.0.0.0 --baseURL http://<your-lan-ip>:1313/ --appendPort=false
```

## Production build

```sh
hugo --gc --minify
```

The static site is written to `docs/public/`.

## Content layout

```
content/        English pages (default language)
content.zh/     Simplified Chinese pages
static/         Images and other static assets
assets/css/     Custom styles (Geist fonts, Pierre syntax theme, pager cards)
layouts/        Theme overrides (home, pager, breadcrumb)
i18n/           UI translations (EN / ZH)
hugo.yaml       Site configuration
```

To add a page, drop a Markdown file into `content/` (and its translation into `content.zh/`). Use `weight:` in the front matter to control sidebar order.
