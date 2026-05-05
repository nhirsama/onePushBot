package admin

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>onePushBot 管理面板</title>
  <style>
    :root {
      color-scheme: light;
      --bg:#f4f7fb;
      --sidebar:#151a26;
      --sidebar-2:#1d2433;
      --card:#ffffff;
      --ink:#1f2937;
      --muted:#6b7280;
      --line:#e5e7eb;
      --primary:#1677ff;
      --primary-soft:#e8f2ff;
      --success:#16a34a;
      --warning:#f59e0b;
      --danger:#dc2626;
      --shadow:0 10px 30px rgba(15,23,42,.08);
    }
    * { box-sizing:border-box; }
    body {
      margin:0;
      min-height:100vh;
      background:var(--bg);
      color:var(--ink);
      font-family:"Segoe UI","Noto Sans SC","PingFang SC",sans-serif;
    }
    button, input, select, textarea { font:inherit; }
    button { border:0; cursor:pointer; }
    .shell { min-height:100vh; display:grid; grid-template-columns:248px 1fr; }
    .sidebar {
      position:sticky; top:0; height:100vh; padding:18px 14px;
      background:linear-gradient(180deg,var(--sidebar),#111622);
      color:#dbe4f0; display:flex; flex-direction:column;
    }
    .brand { display:flex; align-items:center; gap:12px; padding:10px 10px 20px; }
    .logo {
      width:40px; height:40px; border-radius:12px; display:grid; place-items:center;
      background:linear-gradient(135deg,#1683ff,#48c6ef); color:white; font-weight:900;
      box-shadow:0 10px 24px rgba(22,119,255,.32);
    }
    .brand strong { display:block; color:white; font-size:17px; letter-spacing:.01em; }
    .brand span { display:block; color:#8f9bad; font-size:12px; margin-top:2px; }
    .nav { display:grid; gap:6px; margin-top:10px; }
    .nav button {
      width:100%; display:flex; align-items:center; gap:10px; padding:11px 12px;
      border-radius:10px; background:transparent; color:#aab4c3; text-align:left; font-weight:700;
    }
    .nav button:hover { background:rgba(255,255,255,.06); color:white; }
    .nav button.active { background:var(--primary); color:white; box-shadow:0 10px 24px rgba(22,119,255,.24); }
    .nav-mark { width:8px; height:8px; border-radius:50%; background:currentColor; opacity:.75; }
    .side-card {
      margin-top:auto; padding:14px; border-radius:14px; background:rgba(255,255,255,.06);
      border:1px solid rgba(255,255,255,.08); color:#aab4c3; font-size:12px; line-height:1.6;
    }
    .side-card b { color:white; }
    .main { min-width:0; padding:22px 26px 40px; }
    .topbar {
      display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:18px;
    }
    .topbar h1 { margin:0; font-size:26px; letter-spacing:-.03em; }
    .topbar p { margin:6px 0 0; color:var(--muted); font-size:14px; }
    .toolbar { display:flex; flex-wrap:wrap; gap:10px; align-items:center; justify-content:flex-end; }
    .btn {
      min-height:36px; padding:8px 14px; border-radius:9px; font-weight:800;
      color:white; background:var(--primary); box-shadow:0 8px 20px rgba(22,119,255,.18);
    }
    .btn.secondary { background:#334155; box-shadow:none; }
    .btn.ghost { background:white; color:var(--ink); border:1px solid var(--line); box-shadow:none; }
    .btn.warn { background:var(--warning); box-shadow:none; }
    .view { display:none; }
    .view.active { display:block; animation:fadeIn .14s ease-out; }
    @keyframes fadeIn { from { opacity:.55; transform:translateY(3px); } to { opacity:1; transform:none; } }
    .grid { display:grid; gap:14px; }
    .dashboard-grid { grid-template-columns:repeat(4,minmax(0,1fr)); }
    .two-col { grid-template-columns:1fr 1fr; }
    .card {
      background:var(--card); border:1px solid var(--line); border-radius:14px;
      box-shadow:var(--shadow); padding:18px;
    }
    .card h2, .card h3 { margin:0; }
    .card h2 { font-size:18px; }
    .card h3 { font-size:16px; }
    .desc { color:var(--muted); font-size:13px; margin:7px 0 0; }
    .metric { min-height:116px; display:flex; flex-direction:column; justify-content:space-between; }
    .metric label { margin:0; color:var(--muted); font-size:13px; font-weight:700; }
    .metric strong { font-size:26px; letter-spacing:-.04em; margin-top:12px; display:block; }
    .metric .hint { color:var(--muted); font-size:12px; margin-top:8px; }
    .status-row { display:flex; flex-wrap:wrap; gap:8px; margin-top:14px; }
    .pill {
      display:inline-flex; align-items:center; gap:7px; padding:6px 10px; border-radius:999px;
      background:#f8fafc; border:1px solid var(--line); color:var(--muted); font-size:12px; font-weight:800;
    }
    .pill.good { background:#ecfdf3; color:var(--success); border-color:#bbf7d0; }
    .pill.warn { background:#fffbeb; color:#b45309; border-color:#fde68a; }
    .dot { width:7px; height:7px; border-radius:50%; background:currentColor; }
    .fact-list { display:grid; gap:10px; margin-top:14px; }
    .fact {
      display:flex; justify-content:space-between; gap:16px; align-items:center;
      padding:10px 0; border-bottom:1px solid var(--line);
    }
    .fact:last-child { border-bottom:0; }
    .fact b { font-size:13px; }
    .fact span { color:var(--muted); font-size:13px; text-align:right; }
    .platform-card { padding:0; overflow:hidden; }
    .platform-head {
      display:grid; grid-template-columns:1fr auto; gap:16px; align-items:center;
      padding:18px; border-bottom:1px solid var(--line);
    }
    .platform-head h3 { display:flex; align-items:center; gap:10px; }
    .platform-icon {
      width:34px; height:34px; border-radius:10px; display:grid; place-items:center;
      background:var(--primary-soft); color:var(--primary); font-weight:900;
    }
    .platform-body { display:none; padding:6px 18px 18px; background:#fbfdff; }
    .platform-card.enabled { border-color:#b9d8ff; box-shadow:0 10px 28px rgba(22,119,255,.08); }
    .platform-card.enabled .platform-body { display:block; }
    .switch { display:inline-flex; align-items:center; gap:10px; font-weight:800; color:var(--muted); cursor:pointer; }
    .switch input { display:none; }
    .track { width:46px; height:24px; border-radius:999px; background:#cbd5e1; position:relative; transition:.15s ease; }
    .track::after { content:""; width:20px; height:20px; border-radius:50%; background:white; position:absolute; top:2px; left:2px; transition:.15s ease; box-shadow:0 2px 8px rgba(15,23,42,.22); }
    .switch input:checked + .track { background:var(--primary); }
    .switch input:checked + .track::after { transform:translateX(22px); }
    .form-grid { display:grid; grid-template-columns:1fr 1fr; gap:12px 14px; }
    .field.full { grid-column:1 / -1; }
    label { display:block; margin:12px 0 6px; color:#475569; font-size:12px; font-weight:800; }
    input, select, textarea {
      width:100%; border:1px solid #d7dde7; border-radius:9px; background:white; color:var(--ink);
      padding:10px 11px; outline:none;
    }
    input:focus, select:focus, textarea:focus { border-color:var(--primary); box-shadow:0 0 0 3px rgba(22,119,255,.12); }
    textarea { min-height:130px; resize:vertical; font-family:"SFMono-Regular",Consolas,monospace; font-size:13px; }
    .help { color:var(--muted); font-size:12px; margin-top:6px; line-height:1.5; }
    .savebar {
      position:sticky; bottom:14px; z-index:5; margin-top:14px;
      display:flex; align-items:center; justify-content:space-between; gap:12px;
      padding:13px 14px; border:1px solid var(--line); border-radius:14px;
      background:rgba(255,255,255,.92); backdrop-filter:blur(12px); box-shadow:var(--shadow);
    }
    details { margin-top:14px; }
    summary { cursor:pointer; color:var(--muted); font-weight:800; }
    .log-panel { overflow:hidden; background:#0f172a; border:1px solid #1e293b; border-radius:14px; box-shadow:var(--shadow); }
    .log-head {
      display:flex; justify-content:space-between; align-items:center; gap:14px; flex-wrap:wrap;
      padding:16px; color:white; background:#111827; border-bottom:1px solid #1f2937;
    }
    .log-head p { color:#94a3b8; }
    .log-tools { display:flex; flex-wrap:wrap; gap:10px; align-items:center; }
    .log-tools input, .log-tools select { width:auto; min-width:150px; background:#0b1220; color:#e5e7eb; border-color:#334155; }
    .console {
      height:calc(100vh - 210px); min-height:520px; overflow:auto; padding:14px;
      background:#020617; font-family:"SFMono-Regular",Consolas,monospace; font-size:13px; line-height:1.6;
    }
    .log-line {
      display:grid; grid-template-columns:58px 154px 1fr; gap:12px;
      padding:4px 0; border-bottom:1px solid rgba(148,163,184,.10); color:#dbeafe;
    }
    .log-level { font-weight:900; text-transform:uppercase; }
    .log-level.info { color:#86efac; }
    .log-level.warn { color:#facc15; }
    .log-level.error { color:#fb7185; }
    .log-level.debug { color:#7dd3fc; }
    .log-time { color:#94a3b8; }
    .log-empty { color:#94a3b8; padding:28px; text-align:center; }
    .toast {
      position:fixed; right:20px; bottom:20px; padding:12px 16px; border-radius:10px;
      background:#0f172a; color:white; opacity:0; transform:translateY(8px); transition:.16s ease; box-shadow:var(--shadow);
    }
    .toast.show { opacity:1; transform:none; }
    @media (max-width: 980px) {
      .shell { grid-template-columns:1fr; }
      .sidebar { position:relative; height:auto; }
      .nav { grid-template-columns:repeat(3,1fr); }
      .side-card { display:none; }
      .main { padding:18px; }
      .topbar { display:block; }
      .toolbar { justify-content:flex-start; margin-top:12px; }
      .dashboard-grid, .two-col, .form-grid { grid-template-columns:1fr; }
      .field.full { grid-column:auto; }
      .console { height:560px; }
      .log-line { grid-template-columns:1fr; gap:2px; }
    }
  </style>
</head>
<body>
<div class="shell">
  <aside class="sidebar">
    <div class="brand">
      <div class="logo">OP</div>
      <div><strong>onePushBot</strong><span>Web 管理面板</span></div>
    </div>
    <nav class="nav">
      <button class="active" id="tab-dashboard" onclick="showView('dashboard')"><span class="nav-mark"></span>总览</button>
      <button id="tab-platforms" onclick="showView('platforms')"><span class="nav-mark"></span>平台配置</button>
      <button id="tab-logs" onclick="showView('logs')"><span class="nav-mark"></span>运行日志</button>
    </nav>
    <div class="side-card">
      <b>本地安全模式</b><br>
      管理面板默认监听本机地址。Token 会在启动时打印到控制台。
    </div>
  </aside>

  <main class="main">
    <header class="topbar">
      <div>
        <h1 id="pageTitle">总览</h1>
        <p id="pageDesc">查看当前程序状态、资源占用和平台连接情况。</p>
      </div>
      <div class="toolbar">
        <button class="btn ghost" onclick="loadAll()">刷新</button>
        <button class="btn secondary" onclick="post('/api/reload')">重载配置</button>
        <button class="btn warn" onclick="post('/api/restart')">重启程序</button>
      </div>
    </header>

    <section class="view active" id="view-dashboard">
      <div class="grid dashboard-grid">
        <div class="card metric"><label>主程序</label><strong id="runtimeState">-</strong><div class="hint">当前服务状态</div></div>
        <div class="card metric"><label>当前内存</label><strong id="memoryAlloc">-</strong><div class="hint">Go 运行时分配</div></div>
        <div class="card metric"><label>CPU 使用率</label><strong id="cpuPercent">-</strong><div class="hint">启动以来平均值</div></div>
        <div class="card metric"><label>协程数量</label><strong id="goroutines">-</strong><div class="hint">当前 Goroutine</div></div>
      </div>
      <div class="grid two-col" style="margin-top:14px;">
        <div class="card">
          <h2>平台状态</h2>
          <p class="desc">当前已启动的平台客户端。</p>
          <div class="status-row" id="platformSummary"></div>
        </div>
        <div class="card">
          <h2>运行信息</h2>
          <div class="fact-list">
            <div class="fact"><b>已启用平台</b><span id="enabledPlatforms">-</span></div>
            <div class="fact"><b>启动时长</b><span id="uptime">-</span></div>
            <div class="fact"><b>系统内存</b><span id="memorySys">-</span></div>
            <div class="fact"><b>CPU 时间</b><span id="cpuTime">-</span></div>
            <div class="fact"><b>配置存储</b><span>SQLite KV</span></div>
          </div>
          <label>Router Buffer</label>
          <input id="routerBuffer" type="number" placeholder="128">
          <div class="help">修改后到“平台配置”页点击保存，重启程序后生效。</div>
        </div>
      </div>
    </section>

    <section class="view" id="view-platforms">
      <div class="grid">
        <article class="card platform-card" id="card-napcat">
          <div class="platform-head">
            <div>
              <h3><span class="platform-icon">Q</span>NapCat / QQ</h3>
              <p class="desc">接入 QQ 机器人消息和发送能力。</p>
              <div class="status-row"><span class="pill" id="state-napcat"><span class="dot"></span>未启用</span><span class="pill">WebSocket/API</span></div>
            </div>
            <label class="switch"><input id="enableNapCat" type="checkbox" onchange="toggleCard('napcat')"><span class="track"></span>启用</label>
          </div>
          <div class="platform-body">
            <div class="form-grid">
              <div class="field full"><label>NapCat WebSocket/API 地址</label><input id="napcatURL" placeholder="ws://127.0.0.1:3001/ws"><div class="help">支持 127.0.0.1:3001 或完整 ws/http 地址。</div></div>
              <div><label>Access Token</label><input id="napcatToken" placeholder="留空不修改，****** 表示已配置"></div>
              <div><label>机器人 QQ 号 self_id</label><input id="napcatSelfID" placeholder="可留空"></div>
              <div><label>心跳超时秒数</label><input id="napcatHeartbeat" type="number" placeholder="90"></div>
              <div><label>最大重连次数</label><input id="napcatReconnect" type="number" placeholder="0 表示不限"></div>
            </div>
          </div>
        </article>

        <article class="card platform-card" id="card-telegram">
          <div class="platform-head">
            <div>
              <h3><span class="platform-icon">T</span>Telegram</h3>
              <p class="desc">Bot 模式适合群消息采集；用户态用于个人账号能力。</p>
              <div class="status-row"><span class="pill" id="state-telegram"><span class="dot"></span>未启用</span><span class="pill">Bot API / User MTProto</span></div>
            </div>
            <label class="switch"><input id="enableTelegram" type="checkbox" onchange="toggleCard('telegram')"><span class="track"></span>启用</label>
          </div>
          <div class="platform-body">
            <div class="form-grid">
              <div><label>启用模式</label><select id="telegramMode" onchange="toggleTelegramMode()"><option value="bot">Bot</option><option value="user">用户态</option><option value="both">Bot + 用户态</option></select></div>
              <div><label>Bot 状态通知 chat_id</label><input id="tgBotOwner" placeholder="例如 123456789"></div>
              <div class="field full tg-bot"><label>Bot Token</label><input id="tgBotToken" placeholder="留空不修改，****** 表示已配置"><div class="help">群消息总结优先使用 Bot 模式。需要在 BotFather 关闭 privacy mode，并把 bot 拉进群。</div></div>
              <div class="tg-user"><label>API ID</label><input id="tgUserAPIID" type="number"></div>
              <div class="tg-user"><label>API Hash</label><input id="tgUserAPIHash" placeholder="留空不修改"></div>
              <div class="tg-user"><label>Session 文件路径</label><input id="tgUserSession" placeholder="./data/telegram_user.session"></div>
              <div class="tg-user"><label>登录方式</label><select id="tgUserAuthMode"><option value="qr">扫码登录</option><option value="code">验证码登录</option></select></div>
              <div class="field full tg-user"><label>允许来源 chat_id，逗号分隔</label><input id="tgUserSourceChats" placeholder="-100xxxxxxxxxx,123456789"></div>
            </div>
          </div>
        </article>

        <article class="card platform-card" id="card-feishu">
          <div class="platform-head">
            <div>
              <h3><span class="platform-icon">F</span>飞书</h3>
              <p class="desc">通过飞书应用和事件回调接入消息。</p>
              <div class="status-row"><span class="pill" id="state-feishu"><span class="dot"></span>未启用</span><span class="pill">Webhook</span></div>
            </div>
            <label class="switch"><input id="enableFeishu" type="checkbox" onchange="toggleCard('feishu')"><span class="track"></span>启用</label>
          </div>
          <div class="platform-body">
            <div class="form-grid">
              <div><label>App ID</label><input id="feishuAppID"></div>
              <div><label>App Secret</label><input id="feishuSecret" placeholder="留空不修改，****** 表示已配置"></div>
              <div><label>Verification Token</label><input id="feishuVerifyToken" placeholder="留空不修改"></div>
              <div><label>Encrypt Key</label><input id="feishuEncryptKey" placeholder="留空不修改"></div>
              <div><label>HTTP 监听地址</label><input id="feishuHTTP" placeholder=":8080"></div>
              <div><label>Webhook Path</label><input id="feishuPath" placeholder="/feishu/events"></div>
              <div class="field full"><label>接收 ID 类型</label><input id="feishuReceiveIDType" placeholder="chat_id"></div>
            </div>
          </div>
        </article>

        <div class="savebar">
          <span class="help">保存只写入配置；保存并重启会立即按新配置重连平台。</span>
          <div class="toolbar">
            <button class="btn" onclick="saveConfig()">保存</button>
            <button class="btn secondary" onclick="saveConfig(true)">保存并重启</button>
          </div>
        </div>

        <details class="card">
          <summary>高级：查看 / 编辑完整 JSON</summary>
          <textarea id="rawConfig"></textarea>
        </details>
      </div>
    </section>

    <section class="view" id="view-logs">
      <div class="log-panel">
        <div class="log-head">
          <div>
            <h2>日志控制台</h2>
            <p class="desc">显示最近的程序日志，适合排查平台连接和配置重启问题。</p>
          </div>
          <div class="log-tools">
            <select id="logLevel" onchange="renderLogs()"><option value="all">全部级别</option><option value="info">Info</option><option value="warn">Warn</option><option value="error">Error</option><option value="debug">Debug</option></select>
            <input id="logSearch" placeholder="搜索日志" oninput="renderLogs()">
            <label class="switch"><input id="logAutoRefresh" type="checkbox" checked><span class="track"></span>自动刷新</label>
            <button class="btn ghost" onclick="clearLogView()">清屏</button>
            <button class="btn" onclick="fetchLogs(false)">刷新日志</button>
          </div>
        </div>
        <div class="console" id="logConsole"><div class="log-empty">暂无日志</div></div>
      </div>
    </section>
  </main>
</div>
<div class="toast" id="toast"></div>
<script>
const token = new URLSearchParams(location.search).get('token') || localStorage.getItem('adminToken') || prompt('Admin Token');
if (token) localStorage.setItem('adminToken', token);
const headers = { 'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json' };
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
  const res = await fetch(path, { ...options, headers: { ...headers, ...(options.headers || {}) } });
  if (!res.ok) throw new Error((await res.json()).error || res.statusText);
  return res.json();
}
async function post(path) {
  await api(path, { method: 'POST' });
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
  if (name === 'napcat') return platforms.includes('qq') || platforms.includes('napcat');
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
  const ids = { napcat: 'enableNapCat', telegram: 'enableTelegram', feishu: 'enableFeishu' };
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
  consoleEl.innerHTML = filtered.map(item => '<div class="log-line"><span class="log-level '+escapeHTML(item.level)+'">'+escapeHTML(item.level)+'</span><span class="log-time">'+escapeHTML(item.time)+'</span><span>'+escapeHTML(item.message)+'</span></div>').join('');
  if (atBottom) consoleEl.scrollTop = consoleEl.scrollHeight;
}
function clearLogView() {
  logs = [];
  document.getElementById('logConsole').innerHTML = '<div class="log-empty">已清屏，后续日志会继续显示</div>';
}
function escapeHTML(value) {
  return String(value ?? '').replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));
}
async function loadAll() {
  const status = await api('/api/status');
  currentConfig = status.config || await api('/api/config');
  document.getElementById('platformSummary').innerHTML = Object.entries(status.platforms || {}).map(([k,v]) => '<span class="pill good"><span class="dot"></span>'+escapeHTML(k)+': '+escapeHTML(v)+'</span>').join('') || '<span class="pill warn"><span class="dot"></span>当前没有运行中的平台</span>';
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
  document.getElementById('feishuHTTP').value = get(currentConfig, 'feishu.http_addr', ':8080');
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
  let patch = { ...currentConfig };
  try { patch = { ...patch, ...JSON.parse(document.getElementById('rawConfig').value || '{}') }; } catch {}
  const platforms = selectedPlatforms();
  patch.platforms = platforms;
  patch.router = { ...(patch.router || {}), broker_buffer: Number(document.getElementById('routerBuffer').value || 128) };
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
    http_addr: document.getElementById('feishuHTTP').value,
    webhook_path: document.getElementById('feishuPath').value,
    receive_id_type: document.getElementById('feishuReceiveIDType').value
  };
  await api('/api/config', { method: 'PUT', body: JSON.stringify(patch) });
  if (restart) await post('/api/restart'); else await loadAll();
  toast(restart ? '已保存并重启' : '已保存配置');
}
loadAll().catch(err => alert(err.message));
fetchLogs(false).catch(() => {});
setInterval(() => {
  if (document.getElementById('logAutoRefresh')?.checked) fetchLogs(true).catch(() => {});
}, 2000);
</script>
</body>
</html>`
