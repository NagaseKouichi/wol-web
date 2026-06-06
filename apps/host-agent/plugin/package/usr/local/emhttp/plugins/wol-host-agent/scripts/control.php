<?php
$allowed = [
    "Start" => "start",
    "Stop" => "stop",
    "Restart" => "restart",
];

$action = $_POST["action"] ?? "";
if (!isset($allowed[$action])) {
    http_response_code(400);
    echo "Invalid action";
    exit;
}

$command = "/etc/rc.d/rc.wol-host-agent " . escapeshellarg($allowed[$action]);
exec($command . " 2>&1", $output, $retval);

header("Content-Type: text/plain");
echo implode("\n", $output);
if ($retval !== 0) {
    http_response_code(500);
}
?>
