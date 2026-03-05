let localDB = JSON.parse(localStorage.getItem('ant_db_v10')) || {};
let localCFG = JSON.parse(localStorage.getItem('ant_cfg_v10')) || {};

window.onload = () => {
    document.getElementById('doc-fecha').innerText = new Date().toLocaleDateString();
    document.getElementById('doc-ref').innerText = "REF: " + Math.floor(Math.random()*9000 + 1000);
    aplicarConfigV10();
    rellenarSelectV10();
};

function openTab(id, btn) {
    document.querySelectorAll('.tab-content').forEach(p => p.classList.remove('active'));
    document.querySelectorAll('.tab-link').forEach(t => t.classList.remove('active'));
    document.getElementById(id).classList.add('active');
    btn.classList.add('active');
    if(id === 'tab-lista') renderizarBiblioteca();
}

function guardarConfig() {
    localCFG = {
        n: document.getElementById('conf-nombre').value,
        c: document.getElementById('conf-cif').value,
        t: document.getElementById('conf-tel').value,
        d: document.getElementById('conf-dir').value,
        nota: document.getElementById('conf-nota').value,
        l: localCFG.l || ""
    };
    localStorage.setItem('ant_cfg_v10', JSON.stringify(localCFG));
    aplicarConfigV10();
    alert("Configuración guardada correctamente");
}

function subirLogo(el) {
    const r = new FileReader();
    r.onload = e => { localCFG.l = e.target.result; aplicarConfigV10(); };
    r.readAsDataURL(el.files[0]);
}

function aplicarConfigV10() {
    document.getElementById('doc-nombre-emp').innerText = localCFG.n || "EMPRESA";
    document.getElementById('doc-cif').innerText = "CIF: " + (localCFG.c || "---");
    document.getElementById('doc-tel').innerText = "TEL: " + (localCFG.t || "---");
    document.getElementById('doc-dir').innerText = localCFG.d || "---";
    document.getElementById('doc-notas-texto').innerText = localCFG.nota || "Validez: 30 días.";
    
    if(localCFG.l) {
        document.getElementById('logo-img').src = localCFG.l;
        document.getElementById('logo-img').style.display = 'block';
        document.getElementById('logo-placeholder').style.display = 'none';
    }

    // Rellenar inputs config
    document.getElementById('conf-nombre').value = localCFG.n || "";
    document.getElementById('conf-cif').value = localCFG.c || "";
    document.getElementById('conf-tel').value = localCFG.t || "";
    document.getElementById('conf-dir').value = localCFG.d || "";
    document.getElementById('conf-nota').value = localCFG.nota || "";
}

function rellenarSelectV10() {
    const s = document.getElementById('sel-item-v10');
    s.innerHTML = '<option value="">-- Buscar en lista --</option>';
    for(let k in localDB) s.innerHTML += `<option value="${k}">${k}</option>`;
}

function autoFillV10() {
    const val = document.getElementById('sel-item-v10').value;
    if(val) {
        document.getElementById('in-desc-v10').value = val;
        document.getElementById('in-prec-v10').value = localDB[val];
    }
}

function agregarLineaV10() {
    const desc = document.getElementById('in-desc-v10').value;
    const cant = parseFloat(document.getElementById('in-cant-v10').value) || 1;
    let prec = parseFloat(document.getElementById('in-prec-v10').value);
    const tipo = document.getElementById('in-tipo-iva-v10').value;

    if(!desc || isNaN(prec)) return;

    let baseU = (tipo === "con") ? (prec / 1.21) : prec;
    let subtotal = baseU * cant;

    const tr = document.createElement('tr');
    tr.innerHTML = `
        <td>${cant}</td>
        <td>${desc} ${tipo === 'con' ? '<small>(IVA inc.)</small>' : ''}</td>
        <td>${baseU.toFixed(2)}€</td>
        <td class="line-sum">${subtotal.toFixed(2)}€</td>
        <td class="no-print"><span onclick="this.parentElement.parentElement.remove();calcularV10()" style="color:red;cursor:pointer">✖</span></td>
    `;
    document.getElementById('cuerpo-tabla').appendChild(tr);
    calcularV10();
    document.getElementById('in-desc-v10').value = "";
    document.getElementById('in-prec-v10').value = "";
}

function calcularV10() {
    let base = 0;
    document.querySelectorAll('.line-sum').forEach(td => base += parseFloat(td.innerText));
    let iva = base * 0.21;
    document.getElementById('res-base').innerText = base.toFixed(2) + "€";
    document.getElementById('res-iva').innerText = iva.toFixed(2) + "€";
    document.getElementById('res-total').innerText = (base + iva).toFixed(2) + "€";
}

function guardarEnLib() {
    const n = document.getElementById('lib-nombre').value;
    const p = document.getElementById('lib-precio').value;
    if(n && p) {
        localDB[n] = p;
        localStorage.setItem('ant_db_v10', JSON.stringify(localDB));
        renderizarBiblioteca();
        rellenarSelectV10();
        document.getElementById('lib-nombre').value = "";
        document.getElementById('lib-precio').value = "";
    }
}

function renderizarBiblioteca() {
    const cont = document.getElementById('lista-render');
    cont.innerHTML = "";
    for(let k in localDB) {
        cont.innerHTML += `<div class="input-card" style="display:flex; justify-content:space-between">
            <span>${k} (<strong>${localDB[k]}€</strong>)</span>
            <span onclick="borrarItemLib('${k}')" style="color:red; cursor:pointer">✖</span>
        </div>`;
    }
}

function borrarItemLib(k) {
    delete localDB[k];
    localStorage.setItem('ant_db_v10', JSON.stringify(localDB));
    renderizarBiblioteca();
    rellenarSelectV10();
}

function borrarHoja() {
    if(confirm("¿Seguro que quieres borrar todo el presupuesto?")) {
        document.getElementById('cuerpo-tabla').innerHTML = "";
        calcularV10();
    }
}
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
