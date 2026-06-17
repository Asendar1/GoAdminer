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

          <div class="csv-upload">
            <label>Config File</label>
            <div class="csv-upload-zone">
              <input type="file" id="csv-file" accept=".csv">
              <span id="csv-filename">Upload a CSV config file</span>
              <button id="btn-download-csv" class="btn btn-sm btn-outline" type="button">Download Config</button>
            </div>
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
        switchDriver(btn.dataset.driver);
      });
    });

    function switchDriver(driver) {
      toggleBtns.forEach(b => b.classList.remove('active'));
      toggleBtns.forEach(b => { if (b.dataset.driver === driver) b.classList.add('active'); });
      selectedDriver = driver;
      pgFields.style.display = selectedDriver === 'postgres' ? '' : 'none';
      sqliteFields.style.display = selectedDriver === 'sqlite' ? '' : 'none';
      errorEl.style.display = 'none';
    }

    const csvFileInput = document.getElementById('csv-file');
    const csvFilename = document.getElementById('csv-filename');

    document.querySelector('.csv-upload-zone').addEventListener('click', function(e) {
      if (e.target.id !== 'btn-download-csv' && !e.target.closest('#btn-download-csv')) {
        csvFileInput.click();
      }
    });

    csvFileInput.addEventListener('change', function() {
      const file = this.files[0];
      if (!file) return;
      csvFilename.textContent = file.name;
      const reader = new FileReader();
      reader.onload = function(e) {
        try {
          const text = e.target.result;
          const lines = text.split('\n').map(l => l.trim()).filter(l => l.length > 0);
          if (lines.length < 2) throw new Error('CSV must have a header row and a data row');
          if (lines.length > 2) console.warn('CSV has ' + (lines.length - 1) + ' data rows; only the first is used');
          const headers = parseCSVLine(lines[0]);
          const values = parseCSVLine(lines[1]);
          if (headers.length !== values.length) throw new Error('Header/data column mismatch');
          const cfg = {};
          headers.forEach((h, i) => { cfg[h.toLowerCase().trim()] = values[i].trim(); });

          if (cfg.driver === 'sqlite') {
            switchDriver('sqlite');
            if (cfg.filepath) document.getElementById('sqlite-path').value = cfg.filepath;
          } else {
            switchDriver('postgres');
            if (cfg.host) document.getElementById('pg-host').value = cfg.host;
            if (cfg.port) document.getElementById('pg-port').value = cfg.port;
            if (cfg.database) document.getElementById('pg-database').value = cfg.database;
            if (cfg.user) document.getElementById('pg-user').value = cfg.user;
            if (cfg.password) {
              try { document.getElementById('pg-password').value = decodeURIComponent(escape(atob(cfg.password))); }
              catch (_) { document.getElementById('pg-password').value = cfg.password; }
            }
            if (cfg.schema) document.getElementById('pg-schema').value = cfg.schema;
            if (cfg.sslmode) document.getElementById('pg-sslmode').value = cfg.sslmode;
          }
          GoAdminer.showSuccess('Config loaded from ' + file.name);
        } catch (err) {
          GoAdminer.showError('CSV parse error: ' + err.message);
          csvFilename.textContent = 'Upload a CSV config file';
        }
      };
      reader.readAsText(file);
    });

    document.getElementById('btn-download-csv').addEventListener('click', function() {
      let fields, cols;
      if (selectedDriver === 'postgres') {
        cols = ['driver', 'host', 'port', 'database', 'user', 'password', 'schema', 'sslmode'];
        fields = {
          driver: 'postgres',
          host: document.getElementById('pg-host').value,
          port: document.getElementById('pg-port').value,
          database: document.getElementById('pg-database').value,
          user: document.getElementById('pg-user').value,
          password: btoa(unescape(encodeURIComponent(document.getElementById('pg-password').value))),
          schema: document.getElementById('pg-schema').value,
          sslmode: document.getElementById('pg-sslmode').value,
        };
      } else {
        cols = ['driver', 'filepath'];
        fields = {
          driver: 'sqlite',
          filepath: document.getElementById('sqlite-path').value,
        };
      }
      const csv = cols.join(',') + '\n' + cols.map(c => {
        const v = fields[c] || '';
        return v.includes(',') || v.includes('"') || v.includes('\n') ? '"' + v.replace(/"/g, '""') + '"' : v;
      }).join(',');
      const blob = new Blob([csv], { type: 'text/csv' });
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = 'goadminer-config.csv';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(a.href);
      GoAdminer.showSuccess('Config downloaded');
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

  function parseCSVLine(line) {
    const result = [];
    let current = '';
    let inQuotes = false;
    for (let i = 0; i < line.length; i++) {
      const ch = line[i];
      if (inQuotes) {
        if (ch === '"') {
          if (i + 1 < line.length && line[i + 1] === '"') { current += '"'; i++; }
          else { inQuotes = false; }
        } else { current += ch; }
      } else {
        if (ch === '"') { inQuotes = true; }
        else if (ch === ',') { result.push(current); current = ''; }
        else { current += ch; }
      }
    }
    result.push(current);
    return result;
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.connect = { render };
})();
