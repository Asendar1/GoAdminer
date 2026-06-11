(function() {
  function render() {
    const main = document.getElementById('main');
    main.innerHTML = `
      <h1>SQL Query</h1>
      <div class="card">
        <div class="query-editor">
          <textarea id="sql-input" placeholder="Enter SQL statement..."></textarea>
        </div>
        <div class="query-toolbar">
          <button id="btn-execute" class="btn btn-primary">Execute</button>
          <button id="btn-clear" class="btn btn-outline">Clear</button>
          <span id="query-status" style="font-size:12px;color:var(--text-muted)"></span>
        </div>
      </div>
      <div id="query-result" style="margin-top:16px;overflow-x:auto"></div>
    `;

    GoAdminer.setBreadcrumb([
      { label: GoAdminer.state.driver + ' / ' + GoAdminer.state.database, view: 'tables' },
      { label: 'SQL Query', view: 'query' },
    ]);

    const input = document.getElementById('sql-input');
    const executeBtn = document.getElementById('btn-execute');
    const clearBtn = document.getElementById('btn-clear');
    const status = document.getElementById('query-status');
    const result = document.getElementById('query-result');

    input.focus();

    input.addEventListener('keydown', (e) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        executeQuery();
      }
    });

    executeBtn.addEventListener('click', executeQuery);
    clearBtn.addEventListener('click', () => {
      input.value = '';
      result.innerHTML = '';
      status.textContent = '';
      input.focus();
    });

    async function executeQuery() {
      const sql = input.value.trim();
      if (!sql) return;

      executeBtn.disabled = true;
      executeBtn.textContent = 'Running...';
      status.textContent = '';
      result.innerHTML = '';

      try {
        const res = await GoAdminer.api.query(sql);

        if (res.error) {
          result.innerHTML = '<div class="error-msg">' + GoAdminer.escapeHtml(res.error) + '</div>';
          return;
        }

        if (res.columns && res.columns.length > 0) {
          let html = '<table><thead><tr>';
          res.columns.forEach(col => {
            html += '<th>' + GoAdminer.escapeHtml(col) + '</th>';
          });
          html += '</tr></thead><tbody>';

          if (res.rows.length === 0) {
            html += '<tr><td colspan="' + res.columns.length + '"><div class="empty-state">0 rows returned</div></td></tr>';
          } else {
            res.rows.forEach(row => {
              html += '<tr>';
              res.columns.forEach(col => {
                html += '<td>' + GoAdminer.formatValue(row[col]) + '</td>';
              });
              html += '</tr>';
            });
          }

          html += '</tbody></table>';
          result.innerHTML = html;
          status.textContent = res.rows.length + ' row(s) returned';
        } else if (res.affected !== undefined) {
          status.textContent = res.affected + ' row(s) affected';
        } else {
          status.textContent = 'Query executed successfully';
        }
      } catch (err) {
        result.innerHTML = '<div class="error-msg">' + GoAdminer.escapeHtml(err.message) + '</div>';
      } finally {
        executeBtn.disabled = false;
        executeBtn.textContent = 'Execute';
      }
    }
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.query = { render };
})();
