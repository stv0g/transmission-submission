<!--
SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
SPDX-License-Identifier: Apache-2.0
-->

# transmission-submission

`transmission-submission` is a little web service to submit new torrent files and magnet links to [Transmission](https://transmissionbt.com/) via its RPC interface.

## Features

- Push notifications about completed downloads
  - Even when the website has been closed
  - Uses [Push API](https://developer.mozilla.org/en-US/docs/Web/API/Push_API?retiredLocale=de)
- Registration of protocol handler for `magnet:` URI's
  - Uses [`navigator.registerProtocolHandler()`](https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler?retiredLocale=de)

## Author

- Steffen Vogel ([@stv0g](https://github.com/stv0g))
