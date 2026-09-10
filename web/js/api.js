const Auth = (() => {
    const KEY_ACCESS = 'uc.access_token';
    const KEY_REFRESH = 'uc.refresh_token';
    const KEY_USER = 'uc.user';

    function getAccess() { return localStorage.getItem(KEY_ACCESS) || ''; }
    function getRefresh() { return localStorage.getItem(KEY_REFRESH) || ''; }
    function getUser() {
        try { return JSON.parse(localStorage.getItem(KEY_USER) || 'null'); }
        catch (_) { return null; }
    }
    function setSession({ access_token, refresh_token, user }) {
        if (access_token) localStorage.setItem(KEY_ACCESS, access_token);
        if (refresh_token) localStorage.setItem(KEY_REFRESH, refresh_token);
        if (user) localStorage.setItem(KEY_USER, JSON.stringify(user));
    }
    function clear() {
        localStorage.removeItem(KEY_ACCESS);
        localStorage.removeItem(KEY_REFRESH);
        localStorage.removeItem(KEY_USER);
    }
    function isLoggedIn() { return !!getAccess(); }

    function logout() {
        clear();
        window.location.href = '/web/login.html';
    }

    async function refresh() {
        const rt = getRefresh();
        if (!rt) { logout(); return null; }
        const res = await fetch('/api/v1/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: rt }),
        });
        const env = await res.json().catch(() => null);
        if (!res.ok || !env || env.code !== 0) {
            logout();
            return null;
        }
        setSession(env.data);
        return env.data;
    }

    return { getAccess, getRefresh, getUser, setSession, clear, isLoggedIn, logout, refresh };
})();

const Api = (() => {
    async function request(method, path, body, opts = {}) {
        const headers = { 'Content-Type': 'application/json' };
        if (!opts.noAuth && Auth.isLoggedIn()) {
            headers['Authorization'] = 'Bearer ' + Auth.getAccess();
        }
        const init = { method, headers };
        if (body !== undefined) init.body = JSON.stringify(body);

        let res = await fetch(path, init);
        if (res.status === 401 && !opts.noAuth && !opts._retry && Auth.getRefresh()) {
            const fresh = await Auth.refresh();
            if (fresh) {
                headers['Authorization'] = 'Bearer ' + Auth.getAccess();
                res = await fetch(path, { ...init, headers });
            }
        }

        let env;
        try { env = await res.json(); }
        catch (_) { env = { code: res.status, message: res.statusText, data: null }; }

        if (!res.ok || (env && env.code !== 0 && env.code !== undefined)) {
            const err = new Error((env && env.message) || ('HTTP ' + res.status));
            err.code = env && env.code;
            err.status = res.status;
            throw err;
        }
        return env ? env.data : null;
    }

    return {
        get: (path, opts) => request('GET', path, undefined, opts),
        post: (path, body, opts) => request('POST', path, body, opts),
        put: (path, body, opts) => request('PUT', path, body, opts),
        del: (path, opts) => request('DELETE', path, undefined, opts),
    };
})();

const Toast = (() => {
    function show(msg, type = '') {
        const el = document.createElement('div');
        el.className = 'toast' + (type ? ' ' + type : '');
        el.textContent = msg;
        document.body.appendChild(el);
        setTimeout(() => el.remove(), 2500);
    }
    return {
        success: (m) => show(m, 'success'),
        error: (m) => show(m, 'danger'),
        info: (m) => show(m, ''),
    };
})();

function fmtTime(iso) {
    if (!iso) return '';
    const d = new Date(iso);
    if (isNaN(d.getTime())) return iso;
    const pad = (n) => String(n).padStart(2, '0');
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
        + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes());
}

function fmtDate(iso) {
    if (!iso) return '';
    const d = new Date(iso);
    if (isNaN(d.getTime())) return iso;
    const pad = (n) => String(n).padStart(2, '0');
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate());
}

function escapeHtml(s) {
    if (s === undefined || s === null) return '';
    return String(s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

function requireLogin() {
    if (!Auth.isLoggedIn()) {
        window.location.href = '/web/login.html';
        return false;
    }
    return true;
}
