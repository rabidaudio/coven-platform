# The Coven — Community Platform

> Learn. Make. Share. Grow together.

The community platform for **[The Coven](https://atlantacoven.org)** — a collaborative makerspace for the queer community, founded by trans creators and focused on mutual support, skill-sharing, and creative projects in music, cosplay, video, and electronics.


## Repository Layout

This is a large mono-repo covering several different inter-connected projects. The details for each can be found in their sub-project READMEs.

| Path | Description |
| --- | --- |
| [atlantacoven.org/](atlantacoven.org/) | Public website built with [Hugo](https://gohugo.io) — tutorials, docs, project gallery, and landing pages. |
| [member-site/](member-site/) | The backend for logged-in members (Go). |
| [app/](app/) | A cross-platform mobile app in [Flutter](https://flutter.dev) for members. |
| [firmware/](firmware/) | The firmware for the door lock. |
| [scripts/](scripts/) | Helper scripts. |


## Roadmap

The current public site is the first piece of a larger system. See [plan.md](plan.md) for the full architecture, which includes:

- **Public site** (Hugo) — this repo
- **Member portal** (React SPA) — billing, prints, projects, access log
- **Backend API** (FastAPI or Go) — hardware control and Stripe integration
- **Database & Auth** — self-hosted Supabase (PostgreSQL + GoTrue + Storage + Realtime)
- **Edge & SSO** — Pangolin VPS reverse proxy
- **Physical access control** — Raspberry Pi + PN532 NFC + Dormakaba 8375 maglock
- **3D printer & camera monitoring** — PrusaLink polling and MJPEG/RTMP proxying

## Contributing

The Coven is a community space — contributions, corrections, and suggestions are welcome. Open an issue or pull request, or reach out via [hello@atlantacoven.org](mailto:hello@atlantacoven.org).

## License

[MIT](LICENSE) © atlantacoven
