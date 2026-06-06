# WOL Web Android

Android WebView wrapper for the `wol-web` frontend.

## Behavior

- First launch opens a native settings screen.
- The user enters the wol-web domain or IP address and port.
- The app opens wol-web in a WebView after saving the server settings.
- The WebView keeps cookies and DOM storage, so the wol-web login session is
  remembered like it is in a browser.
- A settings button is shown above the WebView. Tap it to change the server
  address later.
- Pull down on the WebView to refresh the current wol-web page.
- HTTP LAN addresses are supported through the app network security config.
- Standard ports are omitted from the loaded URL. For example, HTTPS with port
  `80` loads `https://host/` instead of `https://host:80/`.

## Build

Open `apps/android` in Android Studio and build the `app` module.

Command-line builds require a local Gradle/Android SDK setup:

```bash
cd apps/android
gradle :app:assembleDebug
```

The debug APK will be generated under:

```text
apps/android/app/build/outputs/apk/debug/
```

## Settings

Example for a local wol-web instance:

```text
Domain or IP: 192.168.1.24
Port: 8090
Use HTTPS: off
```

You can also paste a full URL into the domain field, such as:

```text
http://192.168.1.24:8090
```

If wol-web uses HTTPS with a self-signed certificate, Android WebView may block
the page unless the certificate is trusted by the device.
