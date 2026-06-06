package com.nagasekouichi.wolweb;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.SharedPreferences;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.text.InputType;
import android.view.Gravity;
import android.view.ViewGroup;
import android.view.inputmethod.EditorInfo;
import android.webkit.CookieManager;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.Button;
import android.widget.CheckBox;
import android.widget.EditText;
import android.widget.FrameLayout;
import android.widget.ImageButton;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;
import android.widget.Toast;

import androidx.swiperefreshlayout.widget.SwipeRefreshLayout;

public class MainActivity extends Activity {
    private static final String PREFS_NAME = "wol_web_settings";
    private static final String KEY_HOST = "host";
    private static final String KEY_PORT = "port";
    private static final String KEY_HTTPS = "https";

    private SharedPreferences prefs;
    private WebView webView;
    private SwipeRefreshLayout swipeRefreshLayout;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE);
        configureCookies();

        if (hasServerConfig()) {
            showWebView();
        } else {
            showSettings(true);
        }
    }

    private boolean hasServerConfig() {
        return !prefs.getString(KEY_HOST, "").trim().isEmpty()
                && !prefs.getString(KEY_PORT, "").trim().isEmpty();
    }

    private void configureCookies() {
        CookieManager cookieManager = CookieManager.getInstance();
        cookieManager.setAcceptCookie(true);
    }

    @SuppressLint("SetJavaScriptEnabled")
    private void showWebView() {
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setBackgroundColor(getColorValue(R.color.app_background));

        FrameLayout toolbar = new FrameLayout(this);
        toolbar.setBackgroundColor(getColorValue(R.color.app_background));
        root.addView(toolbar, new LinearLayout.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                dp(48)
        ));

        webView = new WebView(this);
        webView.setLayoutParams(new SwipeRefreshLayout.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                ViewGroup.LayoutParams.MATCH_PARENT
        ));

        swipeRefreshLayout = new SwipeRefreshLayout(this);
        swipeRefreshLayout.setColorSchemeResources(R.color.app_primary);
        swipeRefreshLayout.setOnRefreshListener(() -> {
            if (webView != null) {
                webView.reload();
            }
        });
        swipeRefreshLayout.setLayoutParams(new LinearLayout.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                0,
                1
        ));

        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setLoadWithOverviewMode(true);
        settings.setUseWideViewPort(true);
        settings.setMixedContentMode(WebSettings.MIXED_CONTENT_COMPATIBILITY_MODE);

        CookieManager.getInstance().setAcceptThirdPartyCookies(webView, true);
        webView.setWebViewClient(new WebViewClient() {
            @Override
            public void onPageFinished(WebView view, String url) {
                if (swipeRefreshLayout != null) {
                    swipeRefreshLayout.setRefreshing(false);
                }
            }
        });
        webView.setWebChromeClient(new WebChromeClient());

        ImageButton settingsButton = new ImageButton(this);
        settingsButton.setImageResource(R.drawable.ic_settings_24);
        settingsButton.setBackgroundResource(R.drawable.settings_button_bg);
        settingsButton.setContentDescription("Settings");
        settingsButton.setPadding(dp(12), dp(12), dp(12), dp(12));
        settingsButton.setOnClickListener(v -> showSettings(false));

        FrameLayout.LayoutParams buttonParams = new FrameLayout.LayoutParams(dp(44), dp(44));
        buttonParams.gravity = Gravity.CENTER_VERTICAL | Gravity.END;
        buttonParams.setMargins(0, 0, dp(8), 0);

        toolbar.addView(settingsButton, buttonParams);
        swipeRefreshLayout.addView(webView);
        root.addView(swipeRefreshLayout);
        setContentView(root);

        webView.loadUrl(buildServerUrl());
    }

    private void showSettings(boolean firstLaunch) {
        webView = null;
        swipeRefreshLayout = null;

        ScrollView scrollView = new ScrollView(this);
        scrollView.setFillViewport(true);
        scrollView.setBackgroundColor(getColorValue(R.color.app_background));

        LinearLayout container = new LinearLayout(this);
        container.setOrientation(LinearLayout.VERTICAL);
        container.setGravity(Gravity.CENTER_HORIZONTAL);
        container.setPadding(dp(24), dp(40), dp(24), dp(24));
        scrollView.addView(container, new ScrollView.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                ViewGroup.LayoutParams.WRAP_CONTENT
        ));

        TextView title = new TextView(this);
        title.setText("WOL Web Settings");
        title.setTextColor(getColorValue(R.color.app_text));
        title.setTextSize(24);
        title.setGravity(Gravity.CENTER);
        title.setTypeface(null, 1);
        container.addView(title, fullWidthParams(0, 0, 0, 24));

        TextView help = new TextView(this);
        help.setText("Enter the wol-web server address used by this app.");
        help.setTextColor(getColorValue(R.color.app_muted));
        help.setTextSize(14);
        help.setGravity(Gravity.CENTER);
        container.addView(help, fullWidthParams(0, 0, 0, 24));

        EditText hostInput = input("Domain or IP", prefs.getString(KEY_HOST, ""));
        hostInput.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_URI);
        container.addView(label("Domain or IP"));
        container.addView(hostInput, fullWidthParams(0, 6, 0, 16));

        EditText portInput = input("Port", prefs.getString(KEY_PORT, "8090"));
        portInput.setInputType(InputType.TYPE_CLASS_NUMBER);
        portInput.setImeOptions(EditorInfo.IME_ACTION_DONE);
        container.addView(label("Port"));
        container.addView(portInput, fullWidthParams(0, 6, 0, 16));

        CheckBox httpsInput = new CheckBox(this);
        httpsInput.setText("Use HTTPS");
        httpsInput.setTextColor(getColorValue(R.color.app_text));
        httpsInput.setButtonTintList(android.content.res.ColorStateList.valueOf(getColorValue(R.color.app_primary)));
        httpsInput.setChecked(prefs.getBoolean(KEY_HTTPS, false));
        container.addView(httpsInput, fullWidthParams(0, 0, 0, 24));

        Button saveButton = new Button(this);
        saveButton.setText(firstLaunch ? "Save and Open" : "Save");
        saveButton.setTextColor(getColorValue(R.color.app_primary_text));
        saveButton.setBackgroundResource(R.drawable.button_bg);
        saveButton.setAllCaps(false);
        container.addView(saveButton, fullWidthParams(0, 0, 0, firstLaunch ? 0 : 12));

        if (!firstLaunch) {
            Button cancelButton = new Button(this);
            cancelButton.setText("Cancel");
            cancelButton.setTextColor(getColorValue(R.color.app_text));
            cancelButton.setBackgroundColor(Color.TRANSPARENT);
            cancelButton.setAllCaps(false);
            cancelButton.setOnClickListener(v -> showWebView());
            container.addView(cancelButton, fullWidthParams(0, 0, 0, 0));
        }

        saveButton.setOnClickListener(v -> {
            ServerConfig config = normalizeConfig(
                    hostInput.getText().toString(),
                    portInput.getText().toString(),
                    httpsInput.isChecked()
            );

            if (config == null) {
                Toast.makeText(this, "Enter a valid domain/IP and port.", Toast.LENGTH_SHORT).show();
                return;
            }

            prefs.edit()
                    .putString(KEY_HOST, config.host)
                    .putString(KEY_PORT, config.port)
                    .putBoolean(KEY_HTTPS, config.https)
                    .apply();

            showWebView();
        });

        setContentView(scrollView);
    }

    private ServerConfig normalizeConfig(String rawHost, String rawPort, boolean useHttps) {
        String host = rawHost.trim();
        String port = rawPort.trim();
        boolean https = useHttps;

        if (host.startsWith("http://") || host.startsWith("https://")) {
            Uri uri = Uri.parse(host);
            https = "https".equalsIgnoreCase(uri.getScheme());
            host = uri.getHost() == null ? "" : uri.getHost();
            if (uri.getPort() > 0) {
                port = String.valueOf(uri.getPort());
            }
        }

        if (host.isEmpty() || !isValidPort(port)) {
            return null;
        }

        return new ServerConfig(host, port, https);
    }

    private boolean isValidPort(String port) {
        try {
            int value = Integer.parseInt(port);
            return value >= 1 && value <= 65535;
        } catch (NumberFormatException e) {
            return false;
        }
    }

    private String buildServerUrl() {
        String scheme = prefs.getBoolean(KEY_HTTPS, false) ? "https" : "http";
        String host = prefs.getString(KEY_HOST, "").trim();
        String port = prefs.getString(KEY_PORT, "").trim();
        if (isDefaultPort(scheme, port)) {
            return scheme + "://" + host + "/";
        }
        return scheme + "://" + host + ":" + port + "/";
    }

    private boolean isDefaultPort(String scheme, String port) {
        return ("http".equals(scheme) && "80".equals(port))
                || ("https".equals(scheme) && ("80".equals(port) || "443".equals(port)));
    }

    private TextView label(String text) {
        TextView label = new TextView(this);
        label.setText(text);
        label.setTextColor(getColorValue(R.color.app_text));
        label.setTextSize(14);
        label.setTypeface(null, 1);
        return label;
    }

    private EditText input(String hint, String value) {
        EditText input = new EditText(this);
        input.setText(value);
        input.setHint(hint);
        input.setSingleLine(true);
        input.setTextColor(getColorValue(R.color.app_text));
        input.setHintTextColor(getColorValue(R.color.app_muted));
        input.setTextSize(16);
        input.setBackgroundResource(R.drawable.input_bg);
        return input;
    }

    private LinearLayout.LayoutParams fullWidthParams(int left, int top, int right, int bottom) {
        LinearLayout.LayoutParams params = new LinearLayout.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                ViewGroup.LayoutParams.WRAP_CONTENT
        );
        params.setMargins(dp(left), dp(top), dp(right), dp(bottom));
        return params;
    }

    private int dp(int value) {
        return (int) (value * getResources().getDisplayMetrics().density + 0.5f);
    }

    private int getColorValue(int id) {
        return getResources().getColor(id, getTheme());
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
            return;
        }
        super.onBackPressed();
    }

    @Override
    protected void onPause() {
        super.onPause();
        CookieManager.getInstance().flush();
    }

    private static class ServerConfig {
        final String host;
        final String port;
        final boolean https;

        ServerConfig(String host, String port, boolean https) {
            this.host = host;
            this.port = port;
            this.https = https;
        }
    }
}
