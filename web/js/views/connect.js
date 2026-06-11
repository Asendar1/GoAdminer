(function() {
  function render() {
    const main = document.getElementById('main');
    main.innerHTML = `
      <div style="max-width:500px;margin:60px auto">
        <div class="card">
          <h1 style="text-align:center;margin-bottom:8px">GoAdminer</h1>
          <p style="text-align:center;color:var(--text-muted);font-size:13px;margin-bottom:20px">
            Connect to your database
          </p>

          <div class="driver-toggle" id="driver-toggle">
            <button class="active" data-driver="postgres">PostgreSQL</button>
            <button data-driver="sqlite">SQLite</button>
          </div>

          <div id="pg-fields">
            <div class="form-row">
              <div class="form-group">
                <label>Host</label>
                <input type="text" id="pg-host" value="localhost" placeholder="localhost">
              </div>
              <div class="form-group">
                <label>Port</label>
                <input type="number" id="pg-port" value="5432" placeholder="5432">
              </div>
            </div>
            <div class="form-group">
              <label>Database</label>
              <input type="text" id="pg-database" placeholder="mydb">
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>Username</label>
                <input type="text" id="pg-user" placeholder="postgres">
              </div>
              <div class="form-group">
                <label>Password</label>
                <input type="password" id="pg-password" placeholder="">
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>Schema</label>
                <input type="text" id="pg-schema" value="public" placeholder="public">
              </div>
              <div class="form-group">
                <label>SSL Mode</label>
                <select id="pg-sslmode">
                  <option value="disable">disable</option>
                  <option value="require">require</option>
                  <option value="verify-ca">verify-ca</option>
                  <option value="verify-full">verify-full</option>
                </select>
              </div>
            </div>
          </div>

          <div id="sqlite-fields" style="display:none">
            <div class="form-group">
              <label>Database File Path</label>
              <input type="text" id="sqlite-path" placeholder="/data/mydb.sqlite">
              <p style="font-size:12px;color:var(--text-muted);margin-top:4px">
                Path to SQLite file on the server
              </p>
            </div>
          </div>

          <button id="btn-connect" class="btn btn-primary" style="width:100%;margin-top:8px">
            Connect
          </button>
          <p id="connect-error" class="error-msg" style="display:none;margin-top:12px"></p>
        </div>
      </div>
    `;

    const toggleBtns = document.querySelectorAll('#driver-toggle button');
    const pgFields = document.getElementById('pg-fields');
    const sqliteFields = document.getElementById('sqlite-fields');
    const connectBtn = document.getElementById('btn-connect');
    const errorEl = document.getElementById('connect-error');

    let selectedDriver = 'postgres';

    toggleBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        toggleBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        selectedDriver = btn.dataset.driver;
        pgFields.style.display = selectedDriver === 'postgres' ? '' : 'none';
        sqliteFields.style.display = selectedDriver === 'sqlite' ? '' : 'none';
        errorEl.style.display = 'none';
      });
    });

    connectBtn.addEventListener('click', async () => {
      connectBtn.disabled = true;
      connectBtn.textContent = 'Connecting...';
      errorEl.style.display = 'none';

      try {
        let cfg;
        if (selectedDriver === 'postgres') {
          cfg = {
            driver: 'postgres',
            host: document.getElementById('pg-host').value || 'localhost',
            port: parseInt(document.getElementById('pg-port').value) || 5432,
            user: document.getElementById('pg-user').value,
            password: document.getElementById('pg-password').value,
            database: document.getElementById('pg-database').value,
            schema: document.getElementById('pg-schema').value || 'public',
            ssl_mode: document.getElementById('pg-sslmode').value,
          };
        } else {
          cfg = {
            driver: 'sqlite',
            filepath: document.getElementById('sqlite-path').value,
          };
        }

        if (!cfg.database && !cfg.filepath) {
          throw new Error('Database name or file path is required');
        }

        const res = await GoAdminer.api.connect(cfg);
        GoAdminer.setSession(res.session_id, res.driver, cfg.database || cfg.filepath);
        GoAdminer.api.status();
        GoAdminer.navigate('tables');
      } catch (err) {
        errorEl.textContent = err.message;
        errorEl.style.display = '';
      } finally {
        connectBtn.disabled = false;
        connectBtn.textContent = 'Connect';
      }
    });
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.connect = { render };
})();
