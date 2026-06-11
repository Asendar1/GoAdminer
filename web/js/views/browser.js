(function() {
  let currentPage = 1;
  let perPage = 50;
  let sortCol = '';
  let sortOrder = 'asc';
  let searchQuery = '';
  let totalRows = 0;
  let schema = null;

  function render() {
    const table = GoAdminer.currentParams.table || GoAdminer.state.currentTable;
    if (!table) { GoAdminer.navigate('tables'); return; }

    GoAdminer.state.currentTable = table;

    const main = document.getElementById('main');
    main.innerHTML = `
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
        <h1>${GoAdminer.escapeHtml(table)}</h1>
        <div>
          <button id="btn-new-row" class="btn btn-primary btn-sm">+ New Row</button>
          <button id="btn-refresh" class="btn btn-outline btn-sm">Refresh</button>
        </div>
      </div>
      <div class="toolbar">
        <input type="text" id="browser-search" class="search-box" placeholder="Search across all columns...">
        <span id="row-count" style="font-size:13px;color:var(--text-muted)"></span>
      </div>
      <div id="schema-info" style="font-size:12px;color:var(--text-muted);margin-bottom:8px"></div>
      <div id="browser-table-wrap" style="overflow-x:auto">
        <div class="loading">Loading</div>
      </div>
      <div id="browser-pagination" class="pagination"></div>
    `;

    GoAdminer.setBreadcrumb([
      { label: GoAdminer.state.driver + ' / ' + GoAdminer.state.database, view: 'tables' },
      { label: table, view: 'browser', params: { table } },
    ]);

    currentPage = 1;
    sortCol = '';
    sortOrder = 'asc';
    searchQuery = '';

    loadSchema(table).then(() => loadRows(table));

    document.getElementById('browser-search').addEventListener('input', GoAdminer.debounce((e) => {
      searchQuery = e.target.value;
      currentPage = 1;
      loadRows(table);
    }, 300));

    document.getElementById('btn-new-row').addEventListener('click', () => {
      GoAdminer.navigate('editor', { table, mode: 'new' });
    });

    document.getElementById('btn-refresh').addEventListener('click', () => {
      loadRows(table);
    });
  }

  async function loadSchema(table) {
    try {
      schema = await GoAdminer.api.tableSchema(table);
      const info = document.getElementById('schema-info');
      const pkStr = schema.pks.length ? 'PK: ' + schema.pks.join(', ') : 'No PK';
      const fkStr = schema.fks.length ? ' | FK: ' + schema.fks.map(f => f.column + '→' + f.ref_table).join(', ') : '';
      info.textContent = schema.columns.length + ' columns | ' + pkStr + fkStr;
    } catch (err) {
      // schema loading is best-effort
    }
  }

  async function loadRows(table) {
    const wrap = document.getElementById('browser-table-wrap');
    const pagination = document.getElementById('browser-pagination');

    try {
      const params = {
        page: currentPage,
        per_page: perPage,
      };
      if (sortCol) {
        params.sort = sortCol;
        params.order = sortOrder;
      }
      if (searchQuery) {
        params.search = searchQuery;
      }

      const res = await GoAdminer.api.listRows(table, params);
      totalRows = res.total || 0;
      const safeRows = res.rows || [];

      document.getElementById('row-count').textContent = safeRows.length + ' of ' + totalRows + ' rows';

      if (res.columns && res.columns.length) {
        renderTable(res.columns, safeRows, table);
      } else {
        wrap.innerHTML = '<div class="empty-state">No rows</div>';
      }

      renderPagination(table);
    } catch (err) {
      wrap.innerHTML = '<div class="error-msg">' + GoAdminer.escapeHtml(err.message) + '</div>';
    }
  }

  function renderTable(columns, rows, table) {
    const wrap = document.getElementById('browser-table-wrap');
    if (!columns.length) {
      wrap.innerHTML = '<div class="empty-state">No columns found</div>';
      return;
    }

    let html = '<table><thead><tr>';
    columns.forEach(col => {
      const sortIcon = sortCol === col
        ? (sortOrder === 'asc' ? ' ▲' : ' ▼')
        : '';
      html += `<th data-col="${GoAdminer.escapeHtml(col)}">${GoAdminer.escapeHtml(col)}<span class="sort-icon">${sortIcon}</span></th>`;
    });
    html += '<th class="actions">Actions</th>';
    html += '</tr></thead><tbody>';

    if (!rows.length) {
      html += '<tr><td colspan="' + (columns.length + 1) + '"><div class="empty-state">No rows match</div></td></tr>';
    } else {
      rows.forEach((row, idx) => {
        html += '<tr>';
        columns.forEach(col => {
          html += '<td title="' + GoAdminer.escapeHtml(String(row[col] ?? '')) + '">' + GoAdminer.formatValue(row[col]) + '</td>';
        });
        html += '<td class="actions">';
        html += `<button class="btn btn-outline btn-sm btn-edit" data-idx="${idx}">Edit</button>`;
        html += `<button class="btn btn-danger btn-sm btn-delete" data-idx="${idx}">Del</button>`;
        html += '</td></tr>';
      });
    }

    html += '</tbody></table>';
    wrap.innerHTML = html;

    wrap.querySelectorAll('th[data-col]').forEach(th => {
      th.addEventListener('click', () => {
        const col = th.dataset.col;
        if (sortCol === col) {
          sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
        } else {
          sortCol = col;
          sortOrder = 'asc';
        }
        loadRows(table);
      });
    });

    wrap.querySelectorAll('.btn-edit').forEach(btn => {
      btn.addEventListener('click', () => {
        const idx = parseInt(btn.dataset.idx);
        const row = rows[idx];
        GoAdminer.navigate('editor', { table, mode: 'edit', row, schema });
      });
    });

    wrap.querySelectorAll('.btn-delete').forEach(btn => {
      btn.addEventListener('click', async () => {
        const idx = parseInt(btn.dataset.idx);
        const row = rows[idx];
        if (!confirm('Delete this row?')) return;
        try {
          const pk = {};
          (schema?.pks || []).forEach(pkCol => {
            pk[pkCol] = row[pkCol];
          });
          if (Object.keys(pk).length === 0) {
            // fallback: use all columns as pk
            columns.forEach(c => { pk[c] = row[c]; });
          }
          await GoAdminer.api.deleteRow(table, pk);
          GoAdminer.showSuccess('Row deleted');
          loadRows(table);
        } catch (err) {
          GoAdminer.showError('Delete failed: ' + err.message);
        }
      });
    });
  }

  function renderPagination(table) {
    const pagination = document.getElementById('browser-pagination');
    const totalPages = Math.max(1, Math.ceil(totalRows / perPage));

    pagination.innerHTML = `
      <span>Page ${currentPage} of ${totalPages} (${totalRows} total rows)</span>
      <div>
        <button id="page-prev" ${currentPage <= 1 ? 'disabled' : ''}>Prev</button>
        <select id="per-page-select">
          <option value="25" ${perPage === 25 ? 'selected' : ''}>25</option>
          <option value="50" ${perPage === 50 ? 'selected' : ''}>50</option>
          <option value="100" ${perPage === 100 ? 'selected' : ''}>100</option>
          <option value="250" ${perPage === 250 ? 'selected' : ''}>250</option>
        </select>
        <button id="page-next" ${currentPage >= totalPages ? 'disabled' : ''}>Next</button>
      </div>
    `;

    document.getElementById('page-prev').addEventListener('click', () => {
      if (currentPage > 1) { currentPage--; loadRows(table); }
    });

    document.getElementById('page-next').addEventListener('click', () => {
      if (currentPage < totalPages) { currentPage++; loadRows(table); }
    });

    document.getElementById('per-page-select').addEventListener('change', (e) => {
      perPage = parseInt(e.target.value);
      currentPage = 1;
      loadRows(table);
    });
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.browser = { render };
})();
