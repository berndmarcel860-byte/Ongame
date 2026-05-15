<?php
require __DIR__ . '/bootstrap.php';

$action = $_GET['action'] ?? '';
$conn = db();

if ($action === 'register') {
    $payload = read_json_input();
    $email = trim(strtolower($payload['email'] ?? ''));
    $username = trim($payload['username'] ?? '');
    $password = (string)($payload['password'] ?? '');

    if (!filter_var($email, FILTER_VALIDATE_EMAIL) || strlen($username) < 3 || strlen($password) < 8) {
        json_response(['error' => 'invalid registration data'], 422);
    }

    $exists = $conn->prepare('SELECT id FROM users WHERE email = ? OR username = ? LIMIT 1');
    $exists->bind_param('ss', $email, $username);
    $exists->execute();
    if ($exists->get_result()->num_rows > 0) {
        json_response(['error' => 'email or username already exists'], 409);
    }

    $hash = password_hash($password, PASSWORD_DEFAULT);
    $role = 'user';
    $balance = 1000.00;
    $ins = $conn->prepare('INSERT INTO users (email, username, password_hash, role, balance) VALUES (?, ?, ?, ?, ?)');
    $ins->bind_param('ssssd', $email, $username, $hash, $role, $balance);
    $ins->execute();

    json_response(['status' => 'registered']);
}

if ($action === 'login') {
    $payload = read_json_input();
    $email = trim(strtolower($payload['email'] ?? ''));
    $password = (string)($payload['password'] ?? '');

    $stmt = $conn->prepare('SELECT id, email, username, password_hash, role, balance FROM users WHERE email = ? LIMIT 1');
    $stmt->bind_param('s', $email);
    $stmt->execute();
    $user = $stmt->get_result()->fetch_assoc();

    if (!$user || !password_verify($password, $user['password_hash'])) {
        json_response(['error' => 'invalid credentials'], 401);
    }

    $_SESSION['user'] = [
        'id' => (int)$user['id'],
        'email' => $user['email'],
        'username' => $user['username'],
        'role' => $user['role'],
        'balance' => (float)$user['balance'],
    ];

    json_response(['status' => 'logged_in', 'user' => $_SESSION['user']]);
}

if ($action === 'logout') {
    session_destroy();
    json_response(['status' => 'logged_out']);
}

if ($action === 'me') {
    $user = current_user();
    if (!$user) {
        json_response(['user' => null]);
    }
    json_response(['user' => $user]);
}

json_response(['error' => 'unsupported action'], 400);
