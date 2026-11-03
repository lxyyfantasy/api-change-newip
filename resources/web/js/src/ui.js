import "./style.css";
import "./connection";
import * as globals from "./globals";
import { activate, deviceInfo } from './ljw';

const app = document.querySelector('#app');

if (app) {
  app.innerHTML = `
  <div id="connect" style="text-align: center"><table style="display: inline-block">
    <tr><td><span>Host: </span></td><td><input id="host" /></td></tr>
    <tr><td><span>Key: </span></td><td><input id="key" /></td></tr>
  </table></div>
  <div id="password" style="display: none;">
    <input type="password" id="password" />
    <button id="confirm" onclick="confirm()">Confirm</button>
    <button id="cancel" onclick="cancel();">Cancel</button>
  </div>
  <div id="status" style="display: none;">
    <div id="text" style="line-height: 2em"></div>
    <button id="cancel" onclick="cancel();">Cancel</button>
  </div>
  <div id="canvas" style="display: none;">
    <button id="cancel" onclick="cancel();">Cancel</button>
    <canvas id="player"></canvas>
    <canvas id="test-yuv-decoder-canvas"></canvas>
  </div>
  <div id="tools" style="display:none;position:fixed;top:8px;left:8px;z-index:99999;flex-direction:column;gap:6px;">
    <button onclick="showActivate()">激活</button>
    <button onclick="showExpire()">有效期查询</button>
    <button onclick="location.reload()">刷新</button>
  </div>
  <div id="modal" style="display:none;position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:100000;background:#fff;padding:16px;border-radius:8px;box-shadow:0 4px 20px rgba(0,0,0,.3);min-width:320px;"></div>
`;

  let player;
  window.init();

  document.body.onload = () => {
    const host = document.querySelector('#host');
    host.value = localStorage.getItem('custom-rendezvous-server');
    // const id = document.querySelector('#id');
    // id.value = localStorage.getItem('id');
    const key = document.querySelector('#key');
    key.value = localStorage.getItem('key');
    player = YUVCanvas.attach(document.getElementById('player'));
    // globals.sendOffCanvas(document.getElementById('player'));
  };

  // window.connect = () => {
  //   const host = document.querySelector('#host');
  //   localStorage.setItem('custom-rendezvous-server', host.value);
  //   // const id = document.querySelector('#id');
  //   // localStorage.setItem('id', id.value);
  //   const key = document.querySelector('#key');
  //   localStorage.setItem('key', key.value);
  //   const func = async () => {
  //     const conn = globals.newConn();
  //     conn.setMsgbox(msgbox);
  //     conn.setDraw((f) => {
  //       /*
  //       if (!(document.getElementById('player').width > 0)) {
  //         document.getElementById('player').width = f.format.displayWidth;
  //         document.getElementById('player').height = f.format.displayHeight;
  //       }
  //       */
  //       globals.draw(f);
  //       player.drawFrame(f);
  //     });
  //     document.querySelector('div#status').style.display = 'block';
  //     document.querySelector('div#connect').style.display = 'none';
  //     document.querySelector('div#text').innerHTML = 'Connecting ...';
  //     await conn.start(localStorage.getItem('id') || '');
  //   };
  //   func();
  // }
  
  window.connect = () => {
      msgbox('error', 'Error', '禁止使用');
  }

  function msgbox(type, title, text) {
    if (!globals.getConn()) return;
    if (type == 'input-password') {
      document.querySelector('div#status').style.display = 'none';
      document.querySelector('div#password').style.display = 'block';
    } else if (!type) {
      document.querySelector('div#canvas').style.display = 'block';
      document.querySelector('div#password').style.display = 'none';
      document.querySelector('div#status').style.display = 'none';
    } else if (type == 'error') {
      document.querySelector('div#status').style.display = 'block';
      document.querySelector('div#canvas').style.display = 'none';
      document.querySelector('div#text').innerHTML = '<div style="color: red; font-weight: bold;">' + text + '</div>';
    } else {
      document.querySelector('div#password').style.display = 'none';
      document.querySelector('div#status').style.display = 'block';
      document.querySelector('div#text').innerHTML = '<div style="font-weight: bold;">' + text + '</div>';
    }
  }

  window.cancel = () => {
    globals.close();
    document.querySelector('div#connect').style.display = 'block';
    document.querySelector('div#password').style.display = 'none';
    document.querySelector('div#status').style.display = 'none';
    document.querySelector('div#canvas').style.display = 'none';
  }

  window.confirm = () => {
    const password = document.querySelector('input#password').value;
    if (password) {
      document.querySelector('div#password').style.display = 'none';
      globals.getConn().login(password);
    }
  }
  
  if (!globals.isMobile()) {
    document.getElementById('tools').style.display = 'flex';
  }
  
  window.showActivate = () => {
    let account = '';
    try { account = JSON.parse(globals.getByName('option', 'user_info')).name || '' } catch (e) {}
    const m = document.getElementById('modal');
    m.style.display = 'block';
    m.innerHTML = `
      <div style="font-weight:bold;margin-bottom:8px;">激活</div>
      <div style="margin-bottom:6px;">账号：<input id="act-account" value="${account}" readonly style="width:200px;" /></div>
      <div style="margin-bottom:6px;">ID：<input id="act-id" style="width:200px;" /></div>
      <div style="margin-bottom:6px;">激活码：<input id="act-code" style="width:200px;" /></div>
      <div style="margin-bottom:10px;">类别：<select id="act-kind"><option value="lease">开通</option><option value="host">托管</option></select></div>
      <div style="text-align:right;">
        <button onclick="document.getElementById('modal').style.display='none'">取消</button>
        <button onclick="doActivate()">确认</button>
      </div>`;
  }
  
  window.doActivate = async () => {
    const account = document.getElementById('act-account').value;
    const id = document.getElementById('act-id').value.trim();
    const code = document.getElementById('act-code').value.trim();
    const kind = document.getElementById('act-kind').value;
    if (!id || !code) { alert('ID 和激活码必填'); return; }
    const res = await activate(code, kind, id, account);
    alert(res.code === 'OK' ? '激活成功' : (res.msg || '激活失败'));
    if (res.code === 'OK') document.getElementById('modal').style.display = 'none';
  }
  
  window.showExpire = async () => {
    const peers = globals.getPeers();
    const ids = Object.values(peers).map(p => p.info && p.info.id).filter(Boolean);
    const m = document.getElementById('modal');
    m.style.display = 'block';
    m.innerHTML = `<div style="font-weight:bold;margin-bottom:8px;">有效期</div><div id="exp-list">查询中...</div>
      <div style="text-align:right;margin-top:10px;"><button onclick="document.getElementById('modal').style.display='none'">关闭</button></div>`;
    const now = Math.floor(Date.now() / 1000);
    const lines = [];
    for (const id of ids) {
      let r = null;
      try { r = await deviceInfo(id) } catch (e) {}
      const fmt = (t) => {
        if (!r || r.code !== 'OK') return '';
        if (!t) return '未激活';
        const days = Math.ceil((t - now) / 86400);
        if (t < now) return '已到期';
        const txt = days + '天';
        return days <= 3 ? `<span style="color:red">${txt}</span>` : txt;
      };
      lines.push(`<div style="margin:2px 0;">${id}　开通:${fmt(r && r.lease_expire)}　托管:${fmt(r && r.host_expire)}</div>`);
    }
    document.getElementById('exp-list').innerHTML = lines.join('') || '无设备';
  }
}