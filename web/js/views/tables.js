(function() {
  function render() {
    const main = document.getElementById('main');
    main.innerHTML = `
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
        <h1>Tables</h1>
        <span id="table-count" style="font-size:13px;color:var(--text-muted)"></span>
      </div>
      <div class="toolbar">
        <input type="text" id="table-search" class="search-box" placeholder="Search tables...">
      </div>
      <div id="table-list" class="table-list">
        <div class="loading">Loading</div>
      </div>
    `;

    GoAdminer.setBreadcrumb([
      { label: GoAdminer.state.driver + ' / ' + GoAdminer.state.database, view: 'tables' }
    ]);

    loadTables();

    document.getElementById('table-search').addEventListener('input', GoAdminer.debounce(() => {
      filterTables();
    }, 200));
  }

  async function loadTables() {
    try {
      const tables = await GoAdminer.api.listTables(GoAdminer.state.schema);
      GoAdminer.state.tables = tables;
      renderTableList(tables);
    } catch (err) {
      document.getElementById('table-list').innerHTML =
        '<div class="error-msg">' + GoAdminer.escapeHtml(err.message) + '</div>';
    }
  }

  function renderTableList(tables) {
    const container = document.getElementById('table-list');
    document.getElementById('table-count').textContent = tables.length + ' tables';

    if (tables.length === 0) {
      container.innerHTML = '<div class="empty-state">No tables found</div>';
      return;
    }

    container.innerHTML = tables.map(t => `
      <a class="table-item" data-table="${GoAdminer.escapeHtml(t.name)}">
        ${GoAdminer.escapeHtml(t.name)}
        <div class="table-type">${GoAdminer.escapeHtml(t.type)}</div>
      </a>
    `).join('');

    container.querySelectorAll('.table-item').forEach(el => {
      el.addEventListener('click', () => {
        const name = el.dataset.table;
        GoAdminer.state.currentTable = name;
        GoAdminer.navigate('browser', { table: name });
      });
    });
  }

  function filterTables() {
    const q = document.getElementById('table-search').value.toLowerCase();
    const filtered = GoAdminer.state.tables.filter(t =>
      t.name.toLowerCase().includes(q)
    );
    renderTableList(filtered);
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.tables = { render };
})();
