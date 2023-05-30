// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

const writeCookie = (name, value, days) => {
    var date, expires;
    if (days) {
        date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        expires = "; expires=" + date.toGMTString();
    } else {
        expires = "";
    }
    document.cookie = name + "=" + value + expires + "; path=/";
}

const readCookie = (name) => {
    var i, c, ca, nameEQ = name + "=";
    ca = document.cookie.split(';');
    for (i = 0; i < ca.length; i++) {
        c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1, c.length);
        }
        if (c.indexOf(nameEQ) == 0) {
            return c.substring(nameEQ.length, c.length);
        }
    }
    return '';
}

const urlBase64ToUint8Array = (base64String) => {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding)
        .replace(/\-/g, '+')
        .replace(/_/g, '/');

    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);

    for (let i = 0; i < rawData.length; ++i) {
        outputArray[i] = rawData.charCodeAt(i);
    }
    return outputArray;
};

if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/worker.js', {
        scope: '/',
    });
}

const init = async () => {
    const magnetTextArea = document.getElementById("magnets");
    const submitButton = document.getElementById("add");
    const subscriptionInput = document.getElementById("subscription");

    for (let i = 0; i < 2; i++) {
        const registration = await navigator.serviceWorker.ready;
            
        try {
            const subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
            });

            subscriptionInput.value = JSON.stringify(subscription);

            break;
        } catch (error) {
            console.log(error);

            await registration.unregister();
        }
    }

    submitButton.disabled = false;

    const urlSearchParams = new URLSearchParams(window.location.search);
    const params = Object.fromEntries(urlSearchParams.entries());

    if ("magnet" in params) {
        magnetTextArea.innerHTML = params["magnet"];

        submitButton.click();
    }
}

const registerHandler = () => {
    navigator.registerProtocolHandler("magnet", document.location.origin + "/?magnet=%s", "Torrent Submission Handler");
}

const unsubscribe = async () => {
    const registration = await navigator.serviceWorker.ready;
    registration.unregister();
}

init()
