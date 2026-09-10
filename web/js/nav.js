const Nav = (() => {
    const ITEMS = [
        { key: 'home', label: '个人中心', href: '/web/pages/home.html' },
        { key: 'users', label: '用户管理', href: '/web/pages/users.html' },
    ];

    function render(activeKey) {
        const user = Auth.getUser() || {};
        const display = user.nickname || user.username || '我';
        const initial = (display[0] || 'U').toUpperCase();

        const links = ITEMS.map((it) => {
            return `<a href="${it.href}" class="${it.key === activeKey ? 'active' : ''}">${it.label}</a>`;
        }).join('');

        return `
<header class="topbar">
    <div class="brand">User Center</div>
    <nav class="nav">${links}</nav>
    <div class="user">
        <div class="avatar" title="${escapeHtml(display)}">${escapeHtml(initial)}</div>
        <span>${escapeHtml(display)}</span>
        <button class="logout" id="logout-btn">退出</button>
    </div>
</header>`;
    }

    function mount(activeKey) {
        if (!requireLogin()) return false;
        const host = document.getElementById('topbar');
        if (host) {
            host.outerHTML = render(activeKey);
        }
        const btn = document.getElementById('logout-btn');
        if (btn) btn.addEventListener('click', () => {
            if (confirm('确定退出登录吗？')) {
                Api.post('/api/v1/auth/logout', {}).catch(() => {});
                Auth.logout();
            }
        });
        return true;
    }

    async function ensureUserLoaded() {
        if (Auth.getUser()) return Auth.getUser();
        try {
            const me = await Api.get('/api/v1/users/me');
            Auth.setSession({ user: me });
            return me;
        } catch (e) {
            return Auth.getUser() || null;
        }
    }

    return { mount, ensureUserLoaded, ITEMS };
})();
