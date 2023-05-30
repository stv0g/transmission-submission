// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

self.addEventListener('push', (e) => {
    const data = e.data.json();
    self.registration.showNotification(data.title, {
        body: data.body,
        icon: data.icon,
    });
});
