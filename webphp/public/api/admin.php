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

json_response(['error' => 'unsupported action'], 400);
