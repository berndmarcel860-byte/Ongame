<?php
require __DIR__ . '/bootstrap.php';

$user = require_auth();
$action = $_GET['action'] ?? '';
$conn = db();

if ($action === 'profile') {
    $stmt = $conn->prepare('SELECT id, email, username, role, balance, created_at FROM users WHERE id = ? LIMIT 1');
    $stmt->bind_param('i', $user['id']);
    $stmt->execute();
    $profile = $stmt->get_result()->fetch_assoc();
    json_response(['profile' => $profile]);
}

if ($action === 'place-bet') {
    $payload = read_json_input();
    $amount = (float)($payload['betAmount'] ?? 0);
    if ($amount <= 0) {
        json_response(['error' => 'bet amount must be greater than zero'], 422);
    }

    $conn->begin_transaction();
    try {
        $lock = $conn->prepare('SELECT balance FROM users WHERE id = ? FOR UPDATE');
        $lock->bind_param('i', $user['id']);
        $lock->execute();
        $row = $lock->get_result()->fetch_assoc();
        $balance = (float)($row['balance'] ?? 0);

        if ($balance < $amount) {
            $conn->rollback();
            json_response(['error' => 'insufficient balance'], 422);
        }

        $isWin = random_int(1, 100) <= 48 ? 1 : 0;
        $winAmount = $isWin ? round($amount * 1.95, 2) : 0.00;
        $newBalance = $balance - $amount + $winAmount;

        $ins = $conn->prepare('INSERT INTO bets (user_id, bet_amount, win_amount, is_win) VALUES (?, ?, ?, ?)');
        $ins->bind_param('iddi', $user['id'], $amount, $winAmount, $isWin);
        $ins->execute();

        $upd = $conn->prepare('UPDATE users SET balance = ? WHERE id = ?');
        $upd->bind_param('di', $newBalance, $user['id']);
        $upd->execute();

        $_SESSION['user']['balance'] = $newBalance;

        $conn->commit();
        json_response([
            'status' => 'bet_placed',
            'isWin' => (bool)$isWin,
            'winAmount' => $winAmount,
            'balance' => $newBalance,
        ]);
    } catch (Throwable $e) {
        $conn->rollback();
        json_response(['error' => 'failed to place bet'], 500);
    }
}

if ($action === 'history') {
    $stmt = $conn->prepare('SELECT id, bet_amount, win_amount, is_win, created_at FROM bets WHERE user_id = ? ORDER BY id DESC LIMIT 20');
    $stmt->bind_param('i', $user['id']);
    $stmt->execute();
    $rows = [];
    $res = $stmt->get_result();
    while ($item = $res->fetch_assoc()) {
        $rows[] = $item;
    }
    json_response(['history' => $rows]);
}

json_response(['error' => 'unsupported action'], 400);
