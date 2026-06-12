(function() {
  function render() {
    const params = GoAdminer.currentParams;
    const table = params.table;
    const mode = params.mode;
    const rowData = params.row || {};
    const schema = params.schema;

    if (!schema) {
      loadSchemaAndRender(table, mode, rowData);
      return;
    }

    renderForm(table, mode, rowData, schema);
  }

  async function loadSchemaAndRender(table, mode, rowData) {
    try {
      const schema = await GoAdminer.api.tableSchema(table);
      renderForm(table, mode, rowData, schema);
    } catch (err) {
      const main = document.getElementById('main');
      main.innerHTML = '<div class="error-msg">' + GoAdminer.escapeHtml(err.message) + '</div>';
    }
  }

  function isAutoField(col, isEdit) {
    if (col.auto_increment) return true;
    const name = col.name.toLowerCase();
    if (name === 'created_at' || name === 'created' || name === 'updated_at' || name === 'updated') return true;
    if (col.default !== null && !isEdit) return true;
    return false;
  }

  function fieldTags(col) {
    const tags = [];
    if (col.is_pk) tags.push('PK');
    if (col.is_fk) tags.push('FK → ' + (col.fk_ref_table || '?'));
    if (col.auto_increment || col.name.toLowerCase() === 'created_at' || col.name.toLowerCase() === 'created' || col.name.toLowerCase() === 'updated_at' || col.name.toLowerCase() === 'updated') tags.push('auto');
    return tags;
  }

  function renderForm(table, mode, rowData, schema) {
    const isEdit = mode === 'edit';
    const title = isEdit ? 'Edit Row' : 'New Row';

    const main = document.getElementById('main');
    main.innerHTML = `
      <h1>${title}</h1>
      <div class="card" style="max-width:700px">
        <form id="row-form">
          ${schema.columns.map(col => {
            const val = isEdit ? (rowData[col.name] ?? '') : '';
            const readOnly = isAutoField(col, isEdit);
            const tags = fieldTags(col);
            return `
              <div class="form-group">
                <label>
                  ${GoAdminer.escapeHtml(col.name)}
                  <span style="font-weight:400;color:var(--text-muted);font-size:11px">
                    (${GoAdminer.escapeHtml(col.data_type)}${col.nullable ? ', nullable' : ''}${tags.length ? ' | ' + tags.join(', ') : ''})
                  </span>
                </label>
                ${renderInput(col, val, readOnly)}
              </div>
            `;
          }).join('')}
          <div class="modal-actions">
            <button type="button" id="btn-form-cancel" class="btn btn-outline">Cancel</button>
            <button type="submit" id="btn-form-save" class="btn btn-primary">${isEdit ? 'Update' : 'Insert'}</button>
          </div>
        </form>
      </div>
    `;

    GoAdminer.setBreadcrumb([
      { label: GoAdminer.state.driver + ' / ' + GoAdminer.state.database, view: 'tables' },
      { label: table, view: 'browser', params: { table } },
      { label: isEdit ? 'Edit' : 'New', view: 'editor', params: { table, mode } },
    ]);

    document.getElementById('btn-form-cancel').addEventListener('click', () => {
      GoAdminer.navigate('browser', { table });
    });

    document.getElementById('row-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData(e.target);
      const data = {};
      for (const [key, val] of formData.entries()) {
        data[key] = val;
      }

      const saveBtn = document.getElementById('btn-form-save');
      saveBtn.disabled = true;
      saveBtn.textContent = 'Saving...';

      try {
        if (isEdit) {
          const pk = {};
          schema.pks.forEach(pkCol => {
            pk[pkCol] = rowData[pkCol];
          });
          if (Object.keys(pk).length === 0) {
            schema.columns.forEach(c => { pk[c.name] = rowData[c.name]; });
          }
          await GoAdminer.api.updateRow(table, { data, pk });
          GoAdminer.showSuccess('Row updated');
        } else {
          await GoAdminer.api.insertRow(table, data);
          GoAdminer.showSuccess('Row inserted');
        }
        GoAdminer.navigate('browser', { table });
      } catch (err) {
        GoAdminer.showError('Save failed: ' + err.message);
        saveBtn.disabled = false;
        saveBtn.textContent = isEdit ? 'Update' : 'Insert';
      }
    });
  }

  function renderInput(col, val, readOnly) {
    const name = GoAdminer.escapeHtml(col.name);
    const strVal = val !== null && val !== undefined ? String(val) : '';

    if (readOnly) {
      return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" disabled style="background:#f0f0f0;color:var(--text-muted)">`;
    }

    const dt = col.data_type.toLowerCase();

    if (dt.includes('bool')) {
      const checked = strVal === 'true' || strVal === '1' || strVal === 't';
      return `
        <select name="${name}">
          <option value="" ${!strVal ? 'selected' : ''}>NULL</option>
          <option value="true" ${strVal === 'true' ? 'selected' : ''}>true</option>
          <option value="false" ${strVal === 'false' ? 'selected' : ''}>false</option>
        </select>`;
    }

    if (dt.includes('int') || dt.includes('serial') || dt.includes('numeric') || dt.includes('float') || dt.includes('double') || dt.includes('real')) {
      return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="${col.nullable ? 'NULL' : ''}">`;
    }

    if (dt.includes('text') || dt.includes('json') || dt.includes('blob') || dt.includes('clob')) {
      return `<textarea name="${name}" rows="4" placeholder="${col.nullable ? 'NULL' : ''}">${GoAdminer.escapeHtml(strVal)}</textarea>`;
    }

    if (dt.includes('timestamp') || dt.includes('date') || dt.includes('time')) {
      return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="YYYY-MM-DD HH:MM:SS">`;
    }

    return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="${col.nullable ? 'NULL' : ''}">`;
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.editor = { render };
})();
