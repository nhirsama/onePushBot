const token = new URLSearchParams(location.search).get('token') || localStorage.getItem('adminToken') || prompt('Admin Token');
if (token) localStorage.setItem('adminToken', token);
const headers = {'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json'};
let currentConfig = {};
let logs = [];
let lastLogID = 0;
let activeView = 'dashboard';

const pageMeta = {
    dashboard: ['总览', '查看当前程序状态、资源占用和平台连接情况。'],
    platforms: ['平台配置', '按平台启用能力，并填写连接所需的核心配置。'],
    logs: ['运行日志', '查看程序输出、平台连接错误和配置重启结果。']
};

async function api(path, options = {}) {
    const res = await fetch(path, {...options, headers: {...headers, ...(options.headers || {})}});
    if (!res.ok) throw new Error((await res.json()).error || res.statusText);
    return res.json();
}

async function post(path) {
    await api(path, {method: 'POST'});
    await loadAll();
    toast('操作已完成');
}

function get(obj, path, fallback = '') {
    return path.split('.').reduce((acc, key) => acc && acc[key] !== undefined ? acc[key] : undefined, obj) ?? fallback;
}

function showView(name) {
    activeView = name;
    document.querySelectorAll('.view').forEach(el => el.classList.remove('active'));
    document.querySelectorAll('.nav button').forEach(el => el.classList.remove('active'));
    document.getElementById('view-' + name).classList.add('active');
    document.getElementById('tab-' + name).classList.add('active');
    document.getElementById('pageTitle').textContent = pageMeta[name][0];
    document.getElementById('pageDesc').textContent = pageMeta[name][1];
    if (name === 'logs') fetchLogs(false);
}

function activePlatforms(cfg) {
    if (Array.isArray(cfg.platforms)) return cfg.platforms;
    return [];
}

function hasPlatform(cfg, name) {
    const platforms = activePlatforms(cfg);
    if (name === 'napcat') return platforms.includes('qq');
    if (name === 'telegram') return platforms.includes('telegram_bot') || platforms.includes('telegram_user');
    return platforms.includes(name);
}

function setEnabled(name, enabled) {
    document.getElementById('card-' + name).classList.toggle('enabled', enabled);
    const state = document.getElementById('state-' + name);
    state.innerHTML = '<span class="dot"></span>' + (enabled ? '已启用' : '未启用');
    state.className = 'pill ' + (enabled ? 'good' : 'warn');
}

function toggleCard(name) {
    const ids = {napcat: 'enableNapCat', telegram: 'enableTelegram', feishu: 'enableFeishu'};
    const input = document.getElementById(ids[name]);
    setEnabled(name, input.checked);
}

function toggleTelegramMode() {
    const mode = document.getElementById('telegramMode').value;
    document.querySelectorAll('.tg-bot').forEach(el => el.style.display = (mode === 'bot' || mode === 'both') ? '' : 'none');
    document.querySelectorAll('.tg-user').forEach(el => el.style.display = (mode === 'user' || mode === 'both') ? '' : 'none');
}

function toast(message) {
    const el = document.getElementById('toast');
    el.textContent = message;
    el.classList.add('show');
    setTimeout(() => el.classList.remove('show'), 1800);
}

async function fetchLogs(incremental = true) {
    const query = incremental && lastLogID > 0 ? '?since=' + lastLogID + '&limit=500' : '?limit=500';
    const data = await api('/api/logs' + query);
    const items = data.logs || [];
    if (!incremental) logs = [];
    for (const item of items) {
        logs.push(item);
        if (item.id > lastLogID) lastLogID = item.id;
    }
    if (logs.length > 1000) logs = logs.slice(logs.length - 1000);
    renderLogs();
}

function renderLogs() {
    const level = document.getElementById('logLevel')?.value || 'all';
    const keyword = (document.getElementById('logSearch')?.value || '').toLowerCase();
    const consoleEl = document.getElementById('logConsole');
    if (!consoleEl) return;
    const filtered = logs.filter(item => {
        const levelOk = level === 'all' || item.level === level;
        const text = (item.message || '').toLowerCase();
        return levelOk && (!keyword || text.includes(keyword));
    });
    if (filtered.length === 0) {
        consoleEl.innerHTML = '<div class="log-empty">没有匹配的日志</div>';
        return;
    }
    const atBottom = consoleEl.scrollTop + consoleEl.clientHeight >= consoleEl.scrollHeight - 24;
    consoleEl.innerHTML = filtered.map(item => '<div class="log-line"><span class="log-level ' + escapeHTML(item.level) + '">' + escapeHTML(item.level) + '</span><span class="log-time">' + escapeHTML(item.time) + '</span><span>' + escapeHTML(item.message) + '</span></div>').join('');
    if (atBottom) consoleEl.scrollTop = consoleEl.scrollHeight;
}

function clearLogView() {
    logs = [];
    document.getElementById('logConsole').innerHTML = '<div class="log-empty">已清屏，后续日志会继续显示</div>';
}

function escapeHTML(value) {
    return String(value ?? '').replace(/[&<>"']/g, ch => ({
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
    }[ch]));
}

async function loadAll() {
    const status = await api('/api/status');
    currentConfig = status.config || await api('/api/config');
    document.getElementById('platformSummary').innerHTML = Object.entries(status.platforms || {}).map(([k, v]) => '<span class="pill good"><span class="dot"></span>' + escapeHTML(k) + ': ' + escapeHTML(v) + '</span>').join('') || '<span class="pill warn"><span class="dot"></span>当前没有运行中的平台</span>';
    document.getElementById('runtimeState').textContent = status.running ? '运行中' : '未启动';
    document.getElementById('enabledPlatforms').textContent = activePlatforms(currentConfig).join('、') || '未启用';
    document.getElementById('uptime').textContent = get(status, 'process.uptime', '-');
    document.getElementById('memoryAlloc').textContent = get(status, 'process.memory_alloc', '-');
    document.getElementById('memorySys').textContent = get(status, 'process.memory_sys', '-');
    document.getElementById('goroutines').textContent = get(status, 'process.goroutines', '-');
    document.getElementById('cpuPercent').textContent = get(status, 'process.cpu_percent', '-');
    document.getElementById('cpuTime').textContent = get(status, 'process.cpu_time', '-');

    document.getElementById('enableNapCat').checked = hasPlatform(currentConfig, 'napcat');
    document.getElementById('enableTelegram').checked = hasPlatform(currentConfig, 'telegram');
    document.getElementById('enableFeishu').checked = hasPlatform(currentConfig, 'feishu');
    setEnabled('napcat', document.getElementById('enableNapCat').checked);
    setEnabled('telegram', document.getElementById('enableTelegram').checked);
    setEnabled('feishu', document.getElementById('enableFeishu').checked);

    const platforms = activePlatforms(currentConfig);
    const botOn = platforms.includes('telegram_bot');
    const userOn = platforms.includes('telegram_user');
    document.getElementById('telegramMode').value = botOn && userOn ? 'both' : userOn ? 'user' : 'bot';
    toggleTelegramMode();

    document.getElementById('routerBuffer').value = get(currentConfig, 'router.broker_buffer', 128);
    document.getElementById('napcatURL').value = get(currentConfig, 'qq.api_url', '');
    document.getElementById('napcatToken').value = get(currentConfig, 'qq.token', '');
    document.getElementById('napcatSelfID').value = get(currentConfig, 'qq.self_id', '');
    document.getElementById('napcatHeartbeat').value = get(currentConfig, 'qq.heartbeat_timeout', 90);
    document.getElementById('napcatReconnect').value = get(currentConfig, 'qq.reconnect_max_attempts', 0);

    document.getElementById('tgBotToken').value = get(currentConfig, 'telegram_bot.token');
    document.getElementById('tgBotOwner').value = get(currentConfig, 'telegram_bot.owner_chat_id');
    document.getElementById('tgUserAPIID').value = get(currentConfig, 'telegram_user.api_id');
    document.getElementById('tgUserAPIHash').value = get(currentConfig, 'telegram_user.api_hash');
    document.getElementById('tgUserSession').value = get(currentConfig, 'telegram_user.session_path', './data/telegram_user.session');
    document.getElementById('tgUserAuthMode').value = get(currentConfig, 'telegram_user.auth_mode', 'qr');
    document.getElementById('tgUserSourceChats').value = Array.isArray(get(currentConfig, 'telegram_user.source_chats', [])) ? get(currentConfig, 'telegram_user.source_chats', []).join(',') : '';

    document.getElementById('feishuAppID').value = get(currentConfig, 'feishu.app_id');
    document.getElementById('feishuSecret').value = get(currentConfig, 'feishu.app_secret');
    document.getElementById('feishuVerifyToken').value = get(currentConfig, 'feishu.verification_token');
    document.getElementById('feishuEncryptKey').value = get(currentConfig, 'feishu.encrypt_key');
    document.getElementById('feishuPath').value = get(currentConfig, 'feishu.webhook_path', '/feishu/events');
    document.getElementById('feishuReceiveIDType').value = get(currentConfig, 'feishu.receive_id_type', 'chat_id');
    document.getElementById('rawConfig').value = JSON.stringify(currentConfig, null, 2);
}

function selectedPlatforms() {
    const platforms = [];
    if (document.getElementById('enableNapCat').checked) platforms.push('qq');
    if (document.getElementById('enableTelegram').checked) {
        const mode = document.getElementById('telegramMode').value;
        if (mode === 'bot' || mode === 'both') platforms.push('telegram_bot');
        if (mode === 'user' || mode === 'both') platforms.push('telegram_user');
    }
    if (document.getElementById('enableFeishu').checked) platforms.push('feishu');
    return platforms;
}

async function saveConfig(restart = false) {
    let patch = {...currentConfig};
    try {
        patch = {...patch, ...JSON.parse(document.getElementById('rawConfig').value || '{}')};
    } catch {
    }
    patch.platforms = selectedPlatforms();
    patch.router = {
        ...(patch.router || {}),
        broker_buffer: Number(document.getElementById('routerBuffer').value || 128)
    };
    patch.qq = {
        ...(patch.qq || {}),
        api_url: document.getElementById('napcatURL').value,
        token: document.getElementById('napcatToken').value,
        self_id: document.getElementById('napcatSelfID').value,
        heartbeat_timeout: Number(document.getElementById('napcatHeartbeat').value || 90),
        reconnect_max_attempts: Number(document.getElementById('napcatReconnect').value || 0)
    };
    patch.telegram_bot = {
        ...(patch.telegram_bot || {}),
        token: document.getElementById('tgBotToken').value,
        owner_chat_id: document.getElementById('tgBotOwner').value
    };
    patch.telegram_user = {
        ...(patch.telegram_user || {}),
        api_id: Number(document.getElementById('tgUserAPIID').value || 0),
        api_hash: document.getElementById('tgUserAPIHash').value,
        session_path: document.getElementById('tgUserSession').value,
        auth_mode: document.getElementById('tgUserAuthMode').value,
        source_chats: document.getElementById('tgUserSourceChats').value.split(',').map(s => s.trim()).filter(Boolean)
    };
    patch.feishu = {
        ...(patch.feishu || {}),
        app_id: document.getElementById('feishuAppID').value,
        app_secret: document.getElementById('feishuSecret').value,
        verification_token: document.getElementById('feishuVerifyToken').value,
        encrypt_key: document.getElementById('feishuEncryptKey').value,
        webhook_path: document.getElementById('feishuPath').value,
        receive_id_type: document.getElementById('feishuReceiveIDType').value
    };
    await api('/api/config', {method: 'PUT', body: JSON.stringify(patch)});
    if (restart) await post('/api/restart'); else await loadAll();
    toast(restart ? '已保存并重启' : '已保存配置');
}

loadAll().catch(err => alert(err.message));
fetchLogs(false).catch(() => {
});
setInterval(() => {
    if (document.getElementById('logAutoRefresh')?.checked) fetchLogs(true).catch(() => {
    });
}, 2000);
