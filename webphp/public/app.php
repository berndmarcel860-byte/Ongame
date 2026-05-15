<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Ongame Player Frontend</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet">
  <link href="/assets/css/custom.css" rel="stylesheet">
</head>
<body>
<div class="container py-4">
  <div class="d-flex justify-content-between align-items-center mb-4">
    <h1 class="h3 m-0">Player Frontend</h1>
    <a href="/" class="btn btn-outline-light btn-sm">Back to Landing</a>
  </div>

  <div class="row g-3">
    <div class="col-lg-6">
      <div class="card glass p-4 rounded-4 h-100">
        <h5>Register</h5>
        <div class="mb-2"><input id="regEmail" class="form-control" placeholder="Email"></div>
        <div class="mb-2"><input id="regUsername" class="form-control" placeholder="Username"></div>
        <div class="mb-2"><input id="regPassword" type="password" class="form-control" placeholder="Password (min 8)"></div>
        <button class="btn btn-brand text-white" id="btnRegister">Create Account</button>
      </div>
    </div>
    <div class="col-lg-6">
      <div class="card glass p-4 rounded-4 h-100">
        <h5>Login</h5>
        <div class="mb-2"><input id="loginEmail" class="form-control" placeholder="Email" value="player@ongame.local"></div>
        <div class="mb-2"><input id="loginPassword" type="password" class="form-control" placeholder="Password" value="Player@123"></div>
        <div class="d-flex gap-2">
          <button class="btn btn-brand text-white" id="btnLogin">Login</button>
          <button class="btn btn-outline-light" id="btnLogout">Logout</button>
        </div>
      </div>
    </div>
  </div>

  <div class="card glass p-4 rounded-4 mt-3">
    <h5>Gameplay Demo (AJAX)</h5>
    <div class="row g-2 align-items-end">
      <div class="col-md-4"><label class="form-label">Bet Amount</label><input id="betAmount" type="number" class="form-control" value="10"></div>
      <div class="col-md-4"><button class="btn btn-brand text-white w-100" id="btnBet">Place Bet</button></div>
      <div class="col-md-4"><button class="btn btn-outline-light w-100" id="btnRefresh">Refresh Profile + History</button></div>
    </div>
  </div>

  <div class="row g-3 mt-1">
    <div class="col-lg-5">
      <div class="card glass p-4 rounded-4 h-100">
        <h5>Profile</h5>
        <pre id="profile" class="text-white">Not loaded</pre>
      </div>
    </div>
    <div class="col-lg-7">
      <div class="card glass p-4 rounded-4 h-100">
        <h5>Recent Bets</h5>
        <div class="table-responsive">
          <table class="table table-dark table-hover align-middle mb-0">
            <thead><tr><th>ID</th><th>Bet</th><th>Win</th><th>Result</th><th>Time</th></tr></thead>
            <tbody id="historyRows"><tr><td colspan="5" class="text-soft">No data</td></tr></tbody>
          </table>
        </div>
      </div>
    </div>
  </div>

  <div class="card glass p-3 rounded-4 mt-3">
    <strong>Status</strong>
    <div id="status" class="text-soft">Ready.</div>
  </div>
</div>

<script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
<script>
  function escapeHtml(value) {
    return String(value ?? '').replace(/[&<>"']/g, (char) => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    }[char]));
  }

  function setStatus(msg, isError=false){
    $('#status').text(msg).toggleClass('text-danger', isError).toggleClass('text-soft', !isError);
  }

  function api(url, method='GET', data=null){
    return $.ajax({
      url,
      method,
      contentType: 'application/json',
      data: data ? JSON.stringify(data) : null,
      xhrFields: { withCredentials: true }
    });
  }

  function refreshProfile(){
    api('/api/user.php?action=profile').done(res => {
      $('#profile').text(JSON.stringify(res.profile, null, 2));
    }).fail(xhr => setStatus(xhr.responseJSON?.error || 'Profile load failed', true));
  }

  function refreshHistory(){
    api('/api/user.php?action=history').done(res => {
      const rows = res.history || [];
      if (!rows.length) {
        $('#historyRows').html('<tr><td colspan="5" class="text-soft">No bets yet</td></tr>');
        return;
      }
      $('#historyRows').html(rows.map(r => `<tr>
          <td>${escapeHtml(r.id)}</td><td>$${escapeHtml(r.bet_amount)}</td><td>$${escapeHtml(r.win_amount)}</td>
          <td>${r.is_win == 1 ? 'WIN' : 'LOSE'}</td><td>${escapeHtml(r.created_at)}</td>
      </tr>`).join(''));
    });
  }

  $('#btnRegister').on('click', function(){
    api('/api/auth.php?action=register','POST',{
      email: $('#regEmail').val(),
      username: $('#regUsername').val(),
      password: $('#regPassword').val()
    }).done(() => setStatus('Registered. You can now login.'))
      .fail(xhr => setStatus(xhr.responseJSON?.error || 'Registration failed', true));
  });

  $('#btnLogin').on('click', function(){
    api('/api/auth.php?action=login','POST',{
      email: $('#loginEmail').val(),
      password: $('#loginPassword').val()
    }).done(() => { setStatus('Login successful'); refreshProfile(); refreshHistory(); })
      .fail(xhr => setStatus(xhr.responseJSON?.error || 'Login failed', true));
  });

  $('#btnLogout').on('click', function(){
    api('/api/auth.php?action=logout', 'POST').done(() => {
      $('#profile').text('Not loaded');
      $('#historyRows').html('<tr><td colspan="5" class="text-soft">No data</td></tr>');
      setStatus('Logged out');
    });
  });

  $('#btnBet').on('click', function(){
    api('/api/user.php?action=place-bet','POST',{ betAmount: Number($('#betAmount').val()) })
      .done(res => {
        setStatus(res.isWin ? `Win! +$${res.winAmount}` : 'No win this round');
        refreshProfile();
        refreshHistory();
      })
      .fail(xhr => setStatus(xhr.responseJSON?.error || 'Bet failed', true));
  });

  $('#btnRefresh').on('click', function(){
    refreshProfile();
    refreshHistory();
  });
</script>
</body>
</html>
