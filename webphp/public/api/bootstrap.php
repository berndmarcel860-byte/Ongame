<?php
session_start();

function json_response($data, int $status = 200): void {
    http_response_code($status);
    header('Content-Type: application/json');
    echo json_encode($data);
    exit;
}

function db(): mysqli {
    static $conn = null;
    if ($conn instanceof mysqli) {
        return $conn;
    }

    $host = getenv('WEB_DB_HOST') ?: 'mysql';
    $port = (int)(getenv('WEB_DB_PORT') ?: '3306');
    $user = getenv('WEB_DB_USER') ?: '';
    $pass = getenv('WEB_DB_PASSWORD') ?: '';
    $name = getenv('WEB_DB_NAME') ?: '';
    if ($user === '' || $pass === '' || $name === '') {
        json_response(['error' => 'database environment is not configured'], 500);
    }

    $conn = new mysqli($host, $user, $pass, $name, $port);
    if ($conn->connect_errno) {
        json_response(['error' => 'database connection failed'], 500);
    }
    $conn->set_charset('utf8mb4');

    seed_demo_users_if_missing($conn);

    return $conn;
}

function seed_demo_users_if_missing(mysqli $conn): void {
    $users = [
        ['admin@ongame.local', 'admindemo', 'Admin@123', 'admin', 10000.00],
        ['player@ongame.local', 'playerdemo', 'Player@123', 'user', 1500.00],
    ];

    foreach ($users as [$email, $username, $password, $role, $balance]) {
        $stmt = $conn->prepare('SELECT id FROM users WHERE email = ? LIMIT 1');
        $stmt->bind_param('s', $email);
        $stmt->execute();
        $result = $stmt->get_result();
        if ($result && $result->num_rows > 0) {
            continue;
        }

        $passwordHash = password_hash($password, PASSWORD_DEFAULT);
        $insert = $conn->prepare('INSERT INTO users (email, username, password_hash, role, balance) VALUES (?, ?, ?, ?, ?)');
        $insert->bind_param('ssssd', $email, $username, $passwordHash, $role, $balance);
        $insert->execute();
    }

    $seed = $conn->query("SELECT id FROM users WHERE email='player@ongame.local' LIMIT 1");
    $player = $seed ? $seed->fetch_assoc() : null;
    if (!$player) {
        return;
    }
    $playerId = (int)$player['id'];

    $check = $conn->prepare('SELECT id FROM bets WHERE user_id = ? LIMIT 1');
    $check->bind_param('i', $playerId);
    $check->execute();
    $hasBet = $check->get_result();
    if ($hasBet && $hasBet->num_rows > 0) {
        return;
    }

    $betAmount = 20.00;
    $winAmount = 38.00;
    $isWin = 1;
    $insBet = $conn->prepare('INSERT INTO bets (user_id, bet_amount, win_amount, is_win) VALUES (?, ?, ?, ?)');
    $insBet->bind_param('iddi', $playerId, $betAmount, $winAmount, $isWin);
    $insBet->execute();
}

function current_user(): ?array {
    if (empty($_SESSION['user'])) {
        return null;
    }
    return $_SESSION['user'];
}

function require_auth(): array {
    $user = current_user();
    if (!$user) {
        json_response(['error' => 'not authenticated'], 401);
    }
    return $user;
}

function require_admin(): array {
    $user = require_auth();
    if (($user['role'] ?? '') !== 'admin') {
        json_response(['error' => 'admin access required'], 403);
    }
    return $user;
}

function read_json_input(): array {
    $raw = file_get_contents('php://input') ?: '{}';
    $data = json_decode($raw, true);
    return is_array($data) ? $data : [];
}
