// app.js — JavaScript propio mínimo del visor (sin dependencias de terceros).
// Se encarga de: (1) recarga en vivo por SSE conservando la posición de
// lectura, (2) el selector de documento del índice, (3) enlaces profundos
// (documento y sección) vía el hash, y (4) la aprobación en dos pasos con
// acuse junto al botón.
(function () {
  "use strict";

  var STORE_KEY = "rei.view.doc";
  var IDX_KEY = "rei.view.index";
  var smallViewport = window.matchMedia("(max-width: 1080px)");
  var indexCollapsedState = false;
  var approveState = "idle";

  function indexStored() {
    try {
      var v = localStorage.getItem(IDX_KEY);
      if (v !== null) {
        return v === "1";
      }
    } catch (e) {
      /* sin almacenamiento */
    }
    // Por defecto: colapsado en móvil (menú desplegable), abierto en escritorio.
    return smallViewport.matches;
  }

  function setIndex(collapsed, remember) {
    indexCollapsedState = collapsed;
    if (remember) {
      try {
        localStorage.setItem(IDX_KEY, collapsed ? "1" : "0");
      } catch (e) {
        /* sin almacenamiento */
      }
    }
    document.body.classList.toggle("index-collapsed", collapsed);
    var btns = document.querySelectorAll(".index-toggle");
    for (var i = 0; i < btns.length; i++) {
      btns[i].setAttribute("aria-expanded", collapsed ? "false" : "true");
    }
  }

  function setNotice(text) {
    var el = document.getElementById("notice");
    if (el) {
      el.textContent = text || "";
    }
  }

  function setActionStatus(text) {
    var el = document.getElementById("action-status");
    if (el) {
      el.textContent = text || "";
    }
  }

  function storedDoc() {
    try {
      return localStorage.getItem(STORE_KEY);
    } catch (e) {
      return null;
    }
  }

  function storeDoc(id) {
    try {
      localStorage.setItem(STORE_KEY, id);
    } catch (e) {
      /* sin almacenamiento */
    }
  }

  function activeDocId() {
    var reader = document.querySelector(".reader");
    if (!reader) {
      return null;
    }
    var active = reader.querySelector(".doc-body.is-active");
    return active ? active.getAttribute("data-doc") : null;
  }

  // applySelection muestra el documento activo y marca el índice.
  function applySelection() {
    var reader = document.querySelector(".reader");
    if (!reader) {
      return;
    }
    var bodies = reader.querySelectorAll(".doc-body");
    if (!bodies.length) {
      return;
    }
    var wanted = storedDoc();
    var found = false;
    for (var i = 0; i < bodies.length; i++) {
      if (bodies[i].getAttribute("data-doc") === wanted) {
        found = true;
        break;
      }
    }
    if (!found) {
      wanted = bodies[0].getAttribute("data-doc");
    }
    for (var j = 0; j < bodies.length; j++) {
      bodies[j].classList.toggle("is-active", bodies[j].getAttribute("data-doc") === wanted);
    }
    var items = document.querySelectorAll(".doc");
    for (var k = 0; k < items.length; k++) {
      var link = items[k].querySelector("a.nm");
      items[k].classList.toggle("active", !!link && link.getAttribute("data-doc") === wanted);
    }
  }

  function selectDoc(id) {
    storeDoc(id);
    if (window.history && history.replaceState) {
      history.replaceState(null, "", "#doc-" + id);
    }
    applySelection();
    if (smallViewport.matches && !indexCollapsedState) {
      setIndex(true, true);
    }
    var reader = document.querySelector(".reader");
    if (reader) {
      reader.scrollTop = 0;
    }
  }

  // refresh recarga el contenido conservando la posición de lectura.
  function refresh() {
    var content = document.getElementById("content");
    var reader = document.querySelector(".reader");
    var index = document.querySelector(".index");
    var meta = document.querySelector(".meta");
    var rTop = reader ? reader.scrollTop : 0;
    var iTop = index ? index.scrollTop : 0;
    var mTop = meta ? meta.scrollTop : 0;
    fetch("/content", { headers: { "Accept": "text/html" } })
      .then(function (r) { return r.text(); })
      .then(function (html) {
        if (!content) {
          return;
        }
        content.innerHTML = html;
        applySelection();
        var r2 = document.querySelector(".reader");
        var i2 = document.querySelector(".index");
        var m2 = document.querySelector(".meta");
        if (r2) r2.scrollTop = rTop;
        if (i2) i2.scrollTop = iTop;
        if (m2) m2.scrollTop = mTop;
      })
      .catch(function () { /* recarga best-effort */ });
  }

  function resetApprove() {
    var area = document.getElementById("approve-area");
    var btn = document.getElementById("approve");
    var cancel = document.getElementById("approve-cancel");
    approveState = "idle";
    if (area) area.classList.remove("confirming", "approved");
    if (btn) { btn.disabled = false; btn.hidden = false; btn.textContent = "Firmar aprobación"; }
    if (cancel) { cancel.disabled = false; cancel.hidden = true; }
    setActionStatus("");
  }

  function onApproveClick() {
    var area = document.getElementById("approve-area");
    var btn = document.getElementById("approve");
    var cancel = document.getElementById("approve-cancel");
    if (approveState === "idle") {
      approveState = "confirming";
      if (area) area.classList.add("confirming");
      if (btn) btn.textContent = "Confirmar aprobación";
      if (cancel) cancel.hidden = false;
      return;
    }
    doApprove();
  }

  function doApprove() {
    var area = document.getElementById("approve-area");
    var btn = document.getElementById("approve");
    var cancel = document.getElementById("approve-cancel");
    if (btn) btn.disabled = true;
    if (cancel) cancel.disabled = true;
    setActionStatus("Aprobando…");
    fetch("/approve", { method: "POST" })
      .then(function (r) {
        return r.json().then(function (data) { return { ok: r.ok, data: data }; });
      })
      .then(function (res) {
        if (res.ok) {
          approveState = "approved";
          if (area) { area.classList.remove("confirming"); area.classList.add("approved"); }
          if (btn) btn.hidden = true;
          if (cancel) cancel.hidden = true;
          var msg = res.data.message || "Aprobación registrada.";
          setActionStatus(msg);
          setNotice(msg);
        } else {
          resetApprove();
          setActionStatus("No se pudo aprobar: " + (res.data.error || "error"));
        }
      })
      .catch(function () {
        resetApprove();
        setActionStatus("Error de red al aprobar.");
      });
  }

  // applyHash selecciona el documento (o sección) que indique el hash.
  function applyHash() {
    var hash = (window.location.hash || "").replace(/^#/, "");
    if (!hash) {
      return;
    }
    var reader = document.querySelector(".reader");
    if (!reader) {
      return;
    }
    var docMatch = /^doc-(.+)$/.exec(hash);
    if (docMatch) {
      if (reader.querySelector('.doc-body[data-doc="' + docMatch[1] + '"]')) {
        selectDoc(docMatch[1]);
      }
      return;
    }
    // Un id de heading: selecciona su documento y desplázate hasta él.
    var el = document.getElementById(hash);
    if (!el) {
      return;
    }
    var body = el.closest ? el.closest(".doc-body") : null;
    if (!body) {
      return;
    }
    selectDoc(body.getAttribute("data-doc"));
    el.scrollIntoView({ block: "start" });
  }

  // Delegación de eventos: los nodos se recrean tras cada recarga.
  document.addEventListener("click", function (ev) {
    var t = ev.target;
    if (!t || !t.closest) {
      return;
    }
    var link = t.closest(".doc a.nm, .callout-link");
    if (link) {
      var id = link.getAttribute("data-doc");
      if (id) {
        ev.preventDefault();
        selectDoc(id);
      }
      return;
    }
    if (t.closest("#approve")) {
      onApproveClick();
      return;
    }
    if (t.closest("#approve-cancel")) {
      resetApprove();
      return;
    }
    if (t.closest(".index-toggle")) {
      setIndex(!indexCollapsedState, true);
    }
  });

  // Init
  document.body.classList.add("js-viewer");
  setIndex(indexStored(), false);
  applySelection();
  applyHash();
  window.addEventListener("hashchange", applyHash);

  // SSE: el servidor emite `data: reload` al cambiar los documentos.
  try {
    var es = new EventSource("/events");
    es.onmessage = function (ev) {
      if (ev.data === "reload") {
        refresh();
      }
    };
  } catch (e) {
    /* sin SSE la página sigue siendo útil */
  }
})();
