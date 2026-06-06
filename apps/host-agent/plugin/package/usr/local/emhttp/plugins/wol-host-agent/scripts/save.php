<?php
$plugin = "wol-host-agent";
$configDir = "/boot/config/plugins/" . $plugin;
$configFile = $configDir . "/" . $plugin . ".cfg";

function cfg_value($value) {
    return '"' . str_replace(['\\', '"', "\r", "\n"], ['\\\\', '\\"', '', ''], $value) . '"';
}

$enabled = ($_POST["ENABLED"] ?? "yes") === "yes" ? "yes" : "no";
$token = trim($_POST["TOKEN"] ?? "");
$port = trim($_POST["PORT"] ?? "8765");

if (!preg_match('/^[0-9]+$/', $port) || (int)$port < 1 || (int)$port > 65535) {
    http_response_code(400);
    header("Content-Type: text/plain");
    echo "Invalid port. Use a value between 1 and 65535.";
    exit;
}

if (!is_dir($configDir) && !mkdir($configDir, 0700, true)) {
    http_response_code(500);
    header("Content-Type: text/plain");
    echo "Failed to create config directory.";
    exit;
}

$config = "";
$config .= "ENABLED=" . cfg_value($enabled) . "\n";
$config .= "TOKEN=" . cfg_value($token) . "\n";
$config .= "PORT=" . cfg_value($port) . "\n";

if (file_put_contents($configFile, $config) === false) {
    http_response_code(500);
    header("Content-Type: text/plain");
    echo "Failed to write config file.";
    exit;
}
chmod($configFile, 0600);

exec("/usr/local/emhttp/plugins/" . $plugin . "/scripts/apply.sh 2>&1", $output, $retval);

header("Content-Type: text/plain");
echo implode("\n", $output);
if ($retval !== 0) {
    http_response_code(500);
    if (!empty($output)) {
        echo "\n";
    }
    echo "Failed to apply settings.";
    exit;
}

if (!empty($output)) {
    echo "\n";
}
echo "Settings saved.";
?>
