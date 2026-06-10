const GoAdminer = {
  state: {
    sessionID: null,
    driver: null,
    database: null,
    schema: null,
    currentTable: null,
    tables: [],
    schemaCache: {},
  },

  setSession(sid, driver, db, schema) {
    this.state.sessionID = sid;
    this.state.driver = driver;
    this.state.database = db;
    this.state.schema = schema || 'public';
    document.getElementById('conn-status').className = 'status-dot connected';
    const info = document.getElementById('conn-info');
    info.textContent = driver + (db ? ' / ' + db : '');
    document.getElementById('btn-disconnect').style.display = '';
    document.getElementById('btn-query').style.display = '';
  },

  clearSession() {
    this.state.sessionID = null;
    this.state.driver = null;
    this.state.database = null;
    this.state.schema = null;
    this.state.currentTable = null;
    this.state.tables = [];
    this.state.schemaCache = {};
    document.getElementById('conn-status').className = 'status-dot disconnected';
    document.getElementById('conn-info').textContent = '';
    document.getElementById('btn-disconnect').style.display = 'none';
    document.getElementById('btn-query').style.display = 'none';
  },

  navigate(view, params) {
    this.currentView = view;
    this.currentParams = params || {};
    if (typeof this.render === 'function') {
      this.render();
    }
  },

  showError(msg) {
    const main = document.getElementById('main');
    const el = document.createElement('div');
    el.className = 'error-msg';
    el.textContent = msg;
    main.insertBefore(el, main.firstChild);
    setTimeout(() => el.remove(), 5000);
  },

  showSuccess(msg) {
    const main = document.getElementById('main');
    const el = document.createElement('div');
    el.className = 'success-msg';
    el.textContent = msg;
    main.insertBefore(el, main.firstChild);
    setTimeout(() => el.remove(), 3000);
  },

  setBreadcrumb(items) {
    const bc = document.getElementById('breadcrumb');
    bc.innerHTML = items.map((item, i) => {
      if (i === items.length - 1) {
        return `<span>${item.label}</span>`;
      }
      return `<a href="#" data-nav="${item.view}" data-params='${JSON.stringify(item.params || {})}'>${item.label}</a> &rsaquo; `;
    }).join('');
  },

  getSessionID() {
    return this.state.sessionID;
  },

  escapeHtml(str) {
    if (str === null || str === undefined) return '<em>NULL</em>';
    const div = document.createElement('div');
    div.textContent = String(str);
    return div.innerHTML;
  },

  formatValue(val) {
    if (val === null || val === undefined) return '<em>NULL</em>';
    if (typeof val === 'object') return JSON.stringify(val);
    return this.escapeHtml(val);
  },

  debounce(fn, ms) {
    let timer;
    return (...args) => {
      clearTimeout(timer);
      timer = setTimeout(() => fn(...args), ms);
    };
  },
};
