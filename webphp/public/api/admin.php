<?php
require __DIR__ . '/bootstrap.php';

require_admin();
$action = $_GET['action'] ?? '';
$conn = db();

if ($action === 'users') {
    $result = $conn->query("SELECT id, email, username, role, balance, created_at FROM users ORDER BY id DESC LIMIT 50");
    $users = [];
    while ($row = $result->fetch_assoc()) {
        $users[] = $row;
    }
    json_response(['users' => $users]);
}

if ($action === 'stats') {
    $summary = $conn->query(
        "SELECT 
            COUNT(*) AS totalBets,
            SUM(CASE WHEN is_win = 1 THEN 1 ELSE 0 END) AS winRounds,
            COALESCE(SUM(bet_amount), 0) AS totalBetAmount,
            COALESCE(SUM(win_amount), 0) AS totalWinAmount
         FROM bets"
    )->fetch_assoc();

    $players = $conn->query("SELECT COUNT(*) AS totalPlayers FROM users WHERE role='user'")->fetch_assoc();

    $totalBets = (int)($summary['totalBets'] ?? 0);
    $winRounds = (int)($summary['winRounds'] ?? 0);
    $totalBetAmount = (float)($summary['totalBetAmount'] ?? 0);
    $totalWinAmount = (float)($summary['totalWinAmount'] ?? 0);

    $winRatePercent = $totalBets > 0 ? ($winRounds / $totalBets) * 100 : 0;
    $payoutPercent = $totalBetAmount > 0 ? ($totalWinAmount / $totalBetAmount) * 100 : 0;

    json_response([
        'totalPlayers' => (int)($players['totalPlayers'] ?? 0),
        'totalBets' => $totalBets,
        'winRounds' => $winRounds,
        'winRatePercent' => round($winRatePercent, 2),
        'payoutPercent' => round($payoutPercent, 2),
        'houseEdgePercent' => round(100 - $payoutPercent, 2),
    ]);
}

if ($action === 'add-funds') {
    $payload = read_json_input();
    $userId = (int)($payload['userId'] ?? 0);
    $amount = (float)($payload['amount'] ?? 0);

    if ($userId <= 0 || $amount <= 0) {
        json_response(['error' => 'valid userId and amount are required'], 422);
    }

    $conn->begin_transaction();
    try {
        $lock = $conn->prepare('SELECT balance FROM users WHERE id = ? FOR UPDATE');
        $lock->bind_param('i', $userId);
        $lock->execute();
        $row = $lock->get_result()->fetch_assoc();
        if (!$row) {
            $conn->rollback();
            json_response(['error' => 'user not found'], 404);
        }

        $balance = (float)($row['balance'] ?? 0);
        $newBalance = $balance + $amount;

        $upd = $conn->prepare('UPDATE users SET balance = ? WHERE id = ?');
        $upd->bind_param('di', $newBalance, $userId);
        $upd->execute();
        $conn->commit();

        json_response([
            'status' => 'funds_added',
            'userId' => $userId,
            'amount' => $amount,
            'newBalance' => $newBalance,
        ]);
    } catch (Throwable $e) {
        $conn->rollback();
        error_log('admin add-funds failed: ' . $e->getMessage());
        json_response(['error' => 'failed to add funds'], 500);
    }
}

json_response(['error' => 'unsupported action'], 400);
