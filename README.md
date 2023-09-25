# SecretSwirl

## About

Rotate your secrets and live worry free!

# Requirements

- [Golang 1.21.x](https://go.dev/dl/)
- [NodeJS 18.x.x](https://nodejs.org/en/) Recommended to use [nvm](https://github.com/nvm-sh/nvm#installing-and-updating) or [windows-nvm](https://github.com/coreybutler/nvm-windows#installation--upgrades) to manage NodeJS versions.
- [Wails 2.6.x](https://wails.io/docs/gettingstarted/installation#platform-specific-dependencies)
  - Then run `wails doctor` to ensure you have all the correct system-level dependencies installed.

## Live Development

To run in live development mode:

- `wails dev` in the project directory
  - This will run a Vite development server that will provide very fast hot reload of your frontend changes.

\
If you want to develop in a browser and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.
