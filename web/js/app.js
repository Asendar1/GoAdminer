(function() {
  function renderView() {
    const view = GoAdminer.currentView || 'connect';
    const fn = GoAdminer.views[view];
    if (fn && typeof fn.render === 'function') {
      fn.render();
    } else {
      document.getElementById('main').innerHTML = '<div class="error-msg">View not found: ' + view + '</div>';
    }
  }

  GoAdminer.render = renderView;

  function navigate(view, params) {
    GoAdminer.currentView = view;
    GoAdminer.currentParams = params || {};
    renderView();
  }

  GoAdminer.navigate = navigate;

  document.addEventListener('click', (e) => {
    const link = e.target.closest('[data-nav]');
    if (!link) return;
    e.preventDefault();
    const view = link.dataset.nav;
    let params = {};
    try {
      params = JSON.parse(link.dataset.params || '{}');
    } catch (_) {}
    navigate(view, params);
  });

  document.getElementById('btn-disconnect').addEventListener('click', async () => {
    try {
      await GoAdminer.api.disconnect();
    } catch (_) {}
    GoAdminer.clearSession();
    GoAdminer.navigate('connect');
  });

  async function init() {
    try {
      const status = await GoAdminer.api.status();
      if (status.connected) {
        GoAdminer.setSession(
          GoAdminer.getSessionID(),
          status.driver,
          status.database,
          status.schema
        );
        navigate('tables');
        return;
      }
    } catch (_) {}
    navigate('connect');
  }

  init();
})();
