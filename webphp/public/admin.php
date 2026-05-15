<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Ongame Admin Backend</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet">
  <link href="/assets/css/custom.css" rel="stylesheet">
</head>
<body>
<div class="container py-4">
  <div class="d-flex justify-content-between align-items-center mb-4">
    <h1 class="h3 m-0">Admin Backend</h1>
    <a href="/" class="btn btn-outline-light btn-sm">Back to Landing</a>
  </div>

  <div class="card glass p-4 rounded-4 mb-3">
    <h5>Admin Login</h5>
    <p class="text-soft mb-2">Demo: <code>admin@ongame.local</code> / <code>Admin@123</code></p>
    <div class="row g-2">
      <div class="col-md-4"><input id="adminEmail" class="form-control" value="admin@ongame.local"></div>
      <div class="col-md-4"><input id="adminPassword" type="password" class="form-control" value="Admin@123"></div>
      <div class="col-md-4 d-flex gap-2"><button id="btnAdminLogin" class="btn btn-brand text-white w-100">Login</button><button id="btnAdminLogout" class="btn btn-outline-light w-100">Logout</button></div>
    </div>
  </div>

  <div class="row g-3 mb-3">
    <div class="col-md-3"><div class="metric"><small class="text-soft">Players</small><h3 id="mPlayers">-</h3></div></div>
    <div class="col-md-3"><div class="metric"><small class="text-soft">Total Bets</small><h3 id="mBets">-</h3></div></div>
    <div class="col-md-3"><div class="metric"><small class="text-soft">Win Rate %</small><h3 id="mRate">-</h3></div></div>
    <div class="col-md-3"><div class="metric"><small class="text-soft">House Edge %</small><h3 id="mEdge">-</h3></div></div>
  </div>

  <div class="card glass p-4 rounded-4">
    <div class="d-flex justify-content-between align-items-center mb-2">
      <h5 class="m-0">Users</h5>
      <button id="btnRefreshAdmin" class="btn btn-outline-light btn-sm">Refresh</button>
    </div>
    <div class="table-responsive">
      <table class="table table-dark table-hover align-middle mb-0">
        <thead><tr><th>ID</th><th>Email</th><th>Username</th><th>Role</th><th>Balance</th><th>Created</th></tr></thead>
        <tbody id="adminUsers"><tr><td colspan="6" class="text-soft">No data</td></tr></tbody>
      </table>
    </div>
  </div>

  <div class="card glass p-3 rounded-4 mt-3">
    <strong>Status</strong>
    <div id="adminStatus" class="text-soft">Ready.</div>
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
    $('#adminStatus').text(msg).toggleClass('text-danger', isError).toggleClass('text-soft', !isError);
  }
  function api(url, method='GET', data=null){
    return $.ajax({
      url,
      method,
      contentType:'application/json',
      data: data ? JSON.stringify(data) : null,
      xhrFields: { withCredentials: true }
    });
  }

  function loadStats(){
    return api('/api/admin.php?action=stats').done(res => {
      $('#mPlayers').text(res.totalPlayers);
      $('#mBets').text(res.totalBets);
      $('#mRate').text(res.winRatePercent + '%');
      $('#mEdge').text(res.houseEdgePercent + '%');
    });
  }

  function loadUsers(){
    return api('/api/admin.php?action=users').done(res => {
      const rows = res.users || [];
      if (!rows.length) {
        $('#adminUsers').html('<tr><td colspan="6" class="text-soft">No users found</td></tr>');
        return;
      }
      $('#adminUsers').html(rows.map(u => `<tr>
        <td>${escapeHtml(u.id)}</td><td>${escapeHtml(u.email)}</td><td>${escapeHtml(u.username)}</td><td>${escapeHtml(u.role)}</td><td>$${escapeHtml(u.balance)}</td><td>${escapeHtml(u.created_at)}</td>
      </tr>`).join(''));
    });
  }

  $('#btnAdminLogin').on('click', function(){
    api('/api/auth.php?action=login', 'POST', {
      email: $('#adminEmail').val(),
      password: $('#adminPassword').val()
    }).done(() => {
      setStatus('Admin login successful');
      loadStats();
      loadUsers();
    }).fail(xhr => setStatus(xhr.responseJSON?.error || 'Admin login failed', true));
  });

  $('#btnAdminLogout').on('click', function(){
    api('/api/auth.php?action=logout','POST').done(() => setStatus('Logged out'));
  });

  $('#btnRefreshAdmin').on('click', function(){
    $.when(loadStats(), loadUsers()).fail(xhr => setStatus(xhr.responseJSON?.error || 'Refresh failed', true));
  });
</script>
</body>
</html>
