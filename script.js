let db = JSON.parse(localStorage.getItem('ant_db')) || {};
let cfg = JSON.parse(localStorage.getItem('ant_cfg')) || {};

window.onload = () => {
    document.getElementById('v-fec').innerText = new Date().toLocaleDateString();
    document.getElementById('v-ref').innerText = "REF: " + Math.floor(Math.random()*9000 + 1000);
    applyCfg();
    refreshSel();
};

function showPanel(id, btn) {
    document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.getElementById(id).classList.add('active');
    btn.classList.add('active');
    if(id === 'p-tarifas') renderLib();
}

function saveCfg() {
    cfg.n = document.getElementById('c-nom').value;
    cfg.c = document.getElementById('c-cif').value;
    cfg.t = document.getElementById('c-tel').value;
    cfg.d = document.getElementById('c-dir').value;
    cfg.nota = document.getElementById('c-nota').value;
    localStorage.setItem('ant_cfg', JSON.stringify(cfg));
    applyCfg(); 
    alert("¡Configuración guardada!");
}

function upLogo(el) {
    const r = new FileReader();
    r.onload = e => { cfg.l = e.target.result; applyCfg(); };
    r.readAsDataURL(el.files[0]);
}

function applyCfg() {
    document.getElementById('v-emp').innerText = cfg.n || "NOMBRE EMPRESA";
    document.getElementById('v-cif').innerText = "CIF: " + (cfg.c || "---");
    document.getElementById('v-tel').innerText = "TEL: " + (cfg.t || "---");
    document.getElementById('v-dir').innerText = cfg.d || "DIR: ---";
    document.getElementById('v-notas').innerText = cfg.nota || "Validez del presupuesto: 30 días.";
    
    if(cfg.l) {
        const img = document.getElementById('v-logo');
        img.src = cfg.l; img.style.display = 'block';
        document.getElementById('v-nom-text').style.display = 'none';
    }

    document.getElementById('c-nom').value = cfg.n || "";
    document.getElementById('c-cif').value = cfg.c || "";
    document.getElementById('c-tel').value = cfg.t || "";
    document.getElementById('c-dir').value = cfg.d || "";
    document.getElementById('c-nota').value = cfg.nota || "";
}

function saveLib() {
    const n = document.getElementById('lib-n').value, p = document.getElementById('lib-p').value;
    if(n && p) {
        db[n] = p; localStorage.setItem('ant_db', JSON.stringify(db));
        renderLib(); refreshSel();
        document.getElementById('lib-n').value = ""; document.getElementById('lib-p').value = "";
    }
}

function renderLib() {
    const l = document.getElementById('lista-lib'); l.innerHTML = "";
    for(let k in db) {
        l.innerHTML += `<div class="card" style="display:flex; justify-content:space-between; align-items:center;">
            <span>${k} (<strong>${db[k]}€</strong>)</span>
            <button onclick="delLib('${k}')" style="background:none; border:none; color:red; cursor:pointer; font-size:18px;">✖</button>
        </div>`;
    }
}

function delLib(k) { delete db[k]; localStorage.setItem('ant_db', JSON.stringify(db)); renderLib(); refreshSel(); }

function refreshSel() {
    const s = document.getElementById('sel-item');
    s.innerHTML = '<option value="">-- Buscar en biblioteca --</option>';
    for(let k in db) s.innerHTML += `<option value="${k}">${k}</option>`;
}

function autoFill() {
    const v = document.getElementById('sel-item').value;
    if(v) { document.getElementById('in-desc').value = v; document.getElementById('in-prec').value = db[v]; }
}

function addFila() {
    const d = document.getElementById('in-desc').value;
    const c = parseFloat(document.getElementById('in-cant').value) || 1;
    let p = parseFloat(document.getElementById('in-prec').value);
    const tipoIva = document.getElementById('in-tipo-iva').value;

    if(!d || isNaN(p)) return;

    let baseUnitario = (tipoIva === "con") ? (p / 1.21) : p;
    let totalLineaBase = baseUnitario * c;

    const tr = document.createElement('tr');
    tr.innerHTML = `
        <td>${c}</td>
        <td>${d} ${tipoIva === "con" ? '<small style="color:gray">(IVA inc.)</small>' : ''}</td>
        <td>${baseUnitario.toFixed(2)}€</td>
        <td class="sum">${totalLineaBase.toFixed(2)}€</td>
        <td class="no-print"><span onclick="this.parentElement.parentElement.remove();calc()" style="color:red; cursor:pointer;">✖</span></td>
    `;
    document.getElementById('v-body').appendChild(tr);
    calc();
    document.getElementById('in-desc').value = ""; document.getElementById('in-prec').value = "";
}

function calc() {
    let b = 0; document.querySelectorAll('.sum').forEach(td => b += parseFloat(td.innerText));
    const iva = b * 0.21;
    document.getElementById('r-base').innerText = b.toFixed(2) + "€";
    document.getElementById('r-iva').innerText = iva.toFixed(2) + "€";
    document.getElementById('r-total').innerText = (b + iva).toFixed(2) + "€";
}

function limpiarHoja() {
    if(confirm("¿Vaciamos el presupuesto actual?")) {
        document.getElementById('v-body').innerHTML = ""; calc();
    }
}

function exportJSON() {
    const blob = new Blob([JSON.stringify({db, cfg})], {type:'application/json'});
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = "Copia_Antenista.json"; a.click();
}

function importJSON(e) {
    const r = new FileReader();
    r.onload = ev => {
        const data = JSON.parse(ev.target.result);
        localStorage.setItem('ant_db', JSON.stringify(data.db));
        localStorage.setItem('ant_cfg', JSON.stringify(data.cfg));
        location.reload();
    };
    r.readAsText(e.target.files[0]);
}
    document.getElementById('v-cif').innerText = "CIF: " + (cfg.c || "---");
    document.getElementById('v-tel').innerText = "TEL: " + (cfg.t || "---");
    if(cfg.l) {
        const img = document.getElementById('v-logo');
        img.src = cfg.l; img.style.display = 'block';
        document.getElementById('v-nom-text').style.display = 'none';
    }
    document.getElementById('c-nom').value = cfg.n || "";
    document.getElementById('c-cif').value = cfg.c || "";
    document.getElementById('c-tel').value = cfg.t || "";
}

// BIBLIOTECA
function saveLib() {
    const n = document.getElementById('lib-n').value, p = document.getElementById('lib-p').value;
    if(n && p) {
        db[n] = p; localStorage.setItem('ant_db', JSON.stringify(db));
        renderLib(); refreshSel();
        document.getElementById('lib-n').value = ""; document.getElementById('lib-p').value = "";
    }
}

function renderLib() {
    const l = document.getElementById('lista-lib'); l.innerHTML = "";
    for(let k in db) {
        l.innerHTML += `<div class="card" style="display:flex; justify-content:space-between">
            <span>${k} (<strong>${db[k]}€</strong>)</span>
            <span style="color:red; cursor:pointer" onclick="delLib('${k}')">✖</span>
        </div>`;
    }
}

function delLib(k) { delete db[k]; localStorage.setItem('ant_db', JSON.stringify(db)); renderLib(); refreshSel(); }

function refreshSel() {
    const s = document.getElementById('sel-item');
    s.innerHTML = '<option value="">-- Opcional: Elegir de lista --</option>';
    for(let k in db) s.innerHTML += `<option value="${k}">${k}</option>`;
}

function autoFill() {
    const v = document.getElementById('sel-item').value;
    if(v) { document.getElementById('in-desc').value = v; document.getElementById('in-prec').value = db[v]; }
}

// LÓGICA PRESUPUESTO
function addFila() {
    const d = document.getElementById('in-desc').value, 
          c = parseFloat(document.getElementById('in-cant').value) || 1, 
          p = parseFloat(document.getElementById('in-prec').value);
    
    if(!d || isNaN(p)) { alert("Falta descripción o precio"); return; }

    const tr = document.createElement('tr');
    tr.innerHTML = `
        <td>${c}</td>
        <td>${d}</td>
        <td>${p.toFixed(2)}€</td>
        <td class="sum">${(c*p).toFixed(2)}€</td>
        <td class="no-print"><span onclick="this.parentElement.parentElement.remove();calc()" style="color:red;cursor:pointer">✖</span></td>
    `;
    document.getElementById('v-body').appendChild(tr);
    calc();
    document.getElementById('in-desc').value = ""; document.getElementById('in-prec').value = "";
}

function calc() {
    let b = 0; document.querySelectorAll('.sum').forEach(td => b += parseFloat(td.innerText));
    document.getElementById('r-base').innerText = b.toFixed(2) + "€";
    document.getElementById('r-iva').innerText = (b * 0.21).toFixed(2) + "€";
    document.getElementById('r-total').innerText = (b * 1.21).toFixed(2) + "€";
}

function limpiarTodo() {
    if(confirm("¿Borrar todo el presupuesto?")) {
        document.getElementById('v-body').innerHTML = "";
        calc();
    }
}

// SYNC
function exportJSON() {
    const data = JSON.stringify({db, cfg});
    const blob = new Blob([data], {type:'application/json'});
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = "Datos_Antenas.json"; a.click();
}

function importJSON(e) {
    const r = new FileReader();
    r.onload = ev => {
        const data = JSON.parse(ev.target.result);
        localStorage.setItem('ant_db', JSON.stringify(data.db));
        localStorage.setItem('ant_cfg', JSON.stringify(data.cfg));
        location.reload();
    };
    r.readAsText(e.target.files[0]);
}
