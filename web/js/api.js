(function() {
  const BASE = '/api';

  async function request(method, path, body) {
    const opts = {
      method,
      headers: { 'Content-Type': 'application/json' },
    };
    const sid = GoAdminer.getSessionID();
    if (sid) {
      opts.headers['X-Session-ID'] = sid;
    }
    if (body !== undefined) {
      opts.body = JSON.stringify(body);
    }
    const res = await fetch(BASE + path, opts);
    const text = await res.text();
    if (!res.ok) {
      let msg = res.statusText;
      try {
        const err = JSON.parse(text);
        msg = err.error || err.message || msg;
      } catch (_) {
        if (text) msg = text;
      }
      throw new Error(msg);
    }
    if (!text) return null;
    return JSON.parse(text);
  }

  GoAdminer.api = {
    connect: (cfg) => request('POST', '/connect', cfg),
    status: () => request('GET', '/status'),
    disconnect: () => request('POST', '/disconnect'),
    listDatabases: () => request('GET', '/databases'),
    listTables: (schema) => request('GET', '/tables' + (schema ? '?schema=' + encodeURIComponent(schema) : '')),
    tableSchema: (table) => request('GET', '/tables/' + encodeURIComponent(table) + '/schema'),
    listRows: (table, params) => {
      const q = new URLSearchParams(params || {});
      return request('GET', '/tables/' + encodeURIComponent(table) + '/rows?' + q.toString());
    },
    insertRow: (table, data) => request('POST', '/tables/' + encodeURIComponent(table) + '/rows', data),
    updateRow: (table, data) => request('PUT', '/tables/' + encodeURIComponent(table) + '/rows', data),
    deleteRow: (table, pk) => request('DELETE', '/tables/' + encodeURIComponent(table) + '/rows', { pk }),
    query: (sql) => request('POST', '/query', { sql }),
  };
})();
