// app.js — JavaScript propio mínimo del visor (sin dependencias de terceros).
// Se encarga de: (1) recarga en vivo por SSE conservando la posición de
// lectura, (2) el selector de documento del índice y (3) enlaces profundos
// (documento y sección) vía el hash. El visor es de solo lectura.
(function () {
  "use strict";

  var STORE_KEY = "rei.view.doc";
  var IDX_KEY = "rei.view.index";
  var smallViewport = window.matchMedia("(max-width: 1080px)");
  var indexCollapsedState = false;

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
    var changes = t.closest(".changes-toggle");
    if (changes) {
      var body = changes.closest(".doc-body");
      if (body) {
        var on = body.classList.toggle("show-changes");
        changes.setAttribute("aria-pressed", on ? "true" : "false");
      }
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

  // SSE: el servidor emite `data: reload` al cambiar los documentos. Si la
  // conexión se interrumpe, se reconecta con espera creciente (1s → 15s) y se
  // avisa de forma discreta hasta restablecerse.
  var RETRY_MIN = 1000;
  var RETRY_MAX = 15000;
  var es = null;
  var retryDelay = RETRY_MIN;
  var reconnectTimer = null;

  function scheduleReconnect() {
    if (reconnectTimer) {
      return;
    }
    reconnectTimer = setTimeout(function () {
      reconnectTimer = null;
      connectEvents();
    }, retryDelay);
    retryDelay = Math.min(retryDelay * 2, RETRY_MAX);
  }

  function connectEvents() {
    try {
      es = new EventSource("/events");
    } catch (e) {
      scheduleReconnect();
      return;
    }
    es.onopen = function () {
      retryDelay = RETRY_MIN;
      var el = document.getElementById("notice");
      if (el && el.textContent === "Reconectando…") {
        setNotice("");
      }
    };
    es.onmessage = function (ev) {
      if (ev.data === "reload") {
        refresh();
      }
    };
    es.onerror = function () {
      if (es) {
        es.close();
        es = null;
      }
      setNotice("Reconectando…");
      scheduleReconnect();
    };
  }

  connectEvents();
})();
