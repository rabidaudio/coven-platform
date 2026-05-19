# atlantacoven.org

The public site, deployed to **https://atlantacoven.org**.

## Prerequisites

- [Hugo (extended)](https://gohugo.io/installation/) — required for SCSS and image processing.

## Local Development

```bash
cd atlantacoven.org
hugo server -D
```

The site will be available at `http://localhost:1313`. The `-D` flag includes draft content.

## Build

```bash
cd atlantacoven.org
hugo --minify
```

The static output is written to `atlantacoven.org/public/`.

## Content Structure

- [content/](content/) — Markdown content organized by section (about, charter, learning, projects, workshops, etc.).
- [data/](data/) — Structured YAML data backing many pages (FAQ, tour, code of conduct, etc.).
- [layouts/](layouts/) — Hugo templates and partials.
- [assets/](assets/) — CSS, JS, and images processed through Hugo Pipes.
- [static/](static/) — Files served as-is (CNAME, robots.txt, etc.).
