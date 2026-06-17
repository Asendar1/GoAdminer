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
                    (${GoAdminer.escapeHtml(col.data_type)}${col.element_type ? '<' + GoAdminer.escapeHtml(col.element_type) + '>' : ''}${col.nullable ? ', nullable' : ''}${tags.length ? ' | ' + tags.join(', ') : ''})
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

    setupArrayWidgets();

    document.getElementById('row-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData(e.target);
      const data = {};
      for (const [key, val] of formData.entries()) {
        data[key] = val;
      }

      for (const col of schema.columns) {
        const key = col.name;
        if (!(key in data)) continue;

        if (col.data_type === 'ARRAY' && typeof data[key] === 'string') {
          data[key] = data[key].split(',').map(s => s.trim()).filter(s => s !== '');
        }

        if ((col.data_type === 'jsonb' || col.data_type === 'json') && typeof data[key] === 'string') {
          const trimmed = data[key].trim();
          if (trimmed !== '') {
            try {
              JSON.parse(trimmed);
            } catch (_) {
              GoAdminer.showError('Invalid JSON in field "' + key + '": ' + _.message);
              return;
            }
          }
        }
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

    if (readOnly) {
      const strVal = val !== null && val !== undefined
        ? (Array.isArray(val) ? val.join(', ') : (typeof val === 'object' ? JSON.stringify(val) : String(val)))
        : '';
      return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" disabled style="background:#f0f0f0;color:var(--text-muted)">`;
    }

    const dt = col.data_type.toLowerCase();

    if (dt === 'array') {
      const strVal = val !== null && val !== undefined
        ? (Array.isArray(val) ? val.join(', ') : String(val))
        : '';
      return `
        <div class="array-widget" data-col="${name}">
          <div class="array-compact" style="display:flex;gap:4px">
            <input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="${col.nullable ? 'NULL' : 'Comma-separated values'}" style="flex:1">
            <button type="button" class="array-toggle-btn btn btn-sm btn-outline" title="Edit individually">⊕</button>
          </div>
          <div class="array-expanded" style="display:none">
            <div class="array-items"></div>
            <div style="margin-top:4px;display:flex;gap:4px">
              <button type="button" class="array-add-btn btn btn-sm btn-outline">+ Add</button>
              <button type="button" class="array-toggle-btn btn btn-sm btn-outline" title="Back to compact">⊖ Compact</button>
            </div>
          </div>
        </div>`;
    }

    if (dt.includes('bool')) {
      const checked = val === true || val === 'true' || val === '1' || val === 't';
      return `
        <select name="${name}">
          <option value="" ${!val && val !== false && val !== 'false' ? 'selected' : ''}>NULL</option>
          <option value="true" ${checked ? 'selected' : ''}>true</option>
          <option value="false" ${val === false || val === 'false' ? 'selected' : ''}>false</option>
        </select>`;
    }

    if (dt.includes('int') || dt.includes('serial') || dt.includes('numeric') || dt.includes('float') || dt.includes('double') || dt.includes('real')) {
      const strVal = val !== null && val !== undefined ? String(val) : '';
      return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="${col.nullable ? 'NULL' : ''}">`;
    }

    if (dt.includes('text') || dt.includes('json') || dt.includes('blob') || dt.includes('clob')) {
      const strVal = val !== null && val !== undefined
        ? (typeof val === 'object' ? JSON.stringify(val, null, 2) : String(val))
        : '';
      return `<textarea name="${name}" rows="4" placeholder="${col.nullable ? 'NULL' : ''}">${GoAdminer.escapeHtml(strVal)}</textarea>`;
    }

    if (dt.includes('timestamp') || dt.includes('date') || dt.includes('time')) {
      const strVal = val !== null && val !== undefined ? String(val) : '';
      return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="YYYY-MM-DD HH:MM:SS">`;
    }

    const strVal = val !== null && val !== undefined ? String(val) : '';
    return `<input type="text" name="${name}" value="${GoAdminer.escapeHtml(strVal)}" placeholder="${col.nullable ? 'NULL' : ''}">`;
  }

  function setupArrayWidgets() {
    document.querySelectorAll('.array-widget').forEach(widget => {
      const compact = widget.querySelector('.array-compact');
      const expanded = widget.querySelector('.array-expanded');
      const itemsContainer = expanded.querySelector('.array-items');
      const hiddenInput = compact.querySelector('input');

      widget.querySelectorAll('.array-toggle-btn').forEach(btn => {
        btn.addEventListener('click', () => {
          if (compact.style.display !== 'none') {
            const values = hiddenInput.value.split(',').map(s => s.trim()).filter(s => s !== '');
            renderArrayItems(itemsContainer, values, hiddenInput);
            compact.style.display = 'none';
            expanded.style.display = 'block';
          } else {
            syncExpandedToHidden(itemsContainer, hiddenInput);
            compact.style.display = 'flex';
            expanded.style.display = 'none';
          }
        });
      });

      widget.querySelector('.array-add-btn').addEventListener('click', () => {
        addArrayItem(itemsContainer, hiddenInput);
      });
    });
  }

  function renderArrayItems(container, values, hiddenInput) {
    container.innerHTML = '';
    values.forEach((v, i) => {
      const div = document.createElement('div');
      div.style.cssText = 'display:flex;gap:4px;margin-bottom:4px';
      const input = document.createElement('input');
      input.type = 'text';
      input.value = v;
      input.placeholder = 'value';
      input.style.cssText = 'flex:1;padding:6px 8px;border:1px solid var(--border);border-radius:var(--radius);font-size:13px';
      input.addEventListener('input', () => syncExpandedToHidden(container, hiddenInput));
      const removeBtn = document.createElement('button');
      removeBtn.type = 'button';
      removeBtn.textContent = '✕';
      removeBtn.className = 'btn btn-sm btn-outline';
      removeBtn.addEventListener('click', () => {
        div.remove();
        syncExpandedToHidden(container, hiddenInput);
      });
      div.appendChild(input);
      div.appendChild(removeBtn);
      container.appendChild(div);
    });
    syncExpandedToHidden(container, hiddenInput);
  }

  function addArrayItem(container, hiddenInput) {
    const div = document.createElement('div');
    div.style.cssText = 'display:flex;gap:4px;margin-bottom:4px';
    const input = document.createElement('input');
    input.type = 'text';
    input.placeholder = 'value';
    input.style.cssText = 'flex:1;padding:6px 8px;border:1px solid var(--border);border-radius:var(--radius);font-size:13px';
    input.addEventListener('input', () => syncExpandedToHidden(container, hiddenInput));
    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.textContent = '✕';
    removeBtn.className = 'btn btn-sm btn-outline';
    removeBtn.addEventListener('click', () => {
      div.remove();
      syncExpandedToHidden(container, hiddenInput);
    });
    div.appendChild(input);
    div.appendChild(removeBtn);
    container.appendChild(div);
    input.focus();
    syncExpandedToHidden(container, hiddenInput);
  }

  function syncExpandedToHidden(container, hiddenInput) {
    const values = [];
    container.querySelectorAll('input').forEach(inp => {
      const trimmed = inp.value.trim();
      if (trimmed !== '') values.push(trimmed);
    });
    hiddenInput.value = values.join(', ');
  }

  GoAdminer.views = GoAdminer.views || {};
  GoAdminer.views.editor = { render };
})();
