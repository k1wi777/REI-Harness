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

  // captureState guarda la posición de lectura y el estado de la interfaz que
  // debe sobrevivir al reemplazo de #content: scroll del lector, índice y
  // FICHA, diffs abiertos y hash vigente.
  function captureState() {
    var reader = document.querySelector(".reader");
    var index = document.querySelector(".index");
    var meta = document.querySelector(".meta");
    var diffs = [];
    var open = document.querySelectorAll(".doc-body.show-changes");
    for (var i = 0; i < open.length; i++) {
      diffs.push(open[i].getAttribute("data-doc"));
    }
    return {
      reader: reader ? reader.scrollTop : 0,
      index: index ? index.scrollTop : 0,
      meta: meta ? meta.scrollTop : 0,
      diffs: diffs,
      hash: window.location.hash || ""
    };
  }

  // restoreHashPosition desplaza la vista a la sección vigente. El documento
  // activo lo resuelve applySelection a partir de la selección guardada.
  function restoreHashPosition(hash) {
    var id = (hash || "").replace(/^#/, "");
    if (!id || /^doc-/.test(id)) {
      return;
    }
    var el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ block: "start" });
    }
  }

  // restoreState reaplica sobre el contenido nuevo el estado capturado. Es la
  // misma rutina para el refresco normal y para el evento final.
  function restoreState(state) {
    for (var i = 0; i < state.diffs.length; i++) {
      var body = document.querySelector('.doc-body[data-doc="' + state.diffs[i] + '"]');
      if (!body) {
        continue;
      }
      body.classList.add("show-changes");
      var toggle = body.querySelector(".changes-toggle");
      if (toggle) {
        toggle.setAttribute("aria-pressed", "true");
      }
    }
    applySelection();
    var reader = document.querySelector(".reader");
    var index = document.querySelector(".index");
    var meta = document.querySelector(".meta");
    if (reader) reader.scrollTop = state.reader;
    if (index) index.scrollTop = state.index;
    if (meta) meta.scrollTop = state.meta;
    restoreHashPosition(state.hash);
  }

  // refresh recarga el contenido conservando la posición de lectura.
  function refresh() {
    var content = document.getElementById("content");
    if (!content) {
      return;
    }
    var state = captureState();
    fetch("/content", { headers: { "Accept": "text/html" } })
      .then(function (r) { return r.text(); })
      .then(function (html) {
        content.innerHTML = html;
        restoreState(state);
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

  // SSE: el servidor emite `data: reload` al cambiar los documentos, el estado
  // o los enlaces. Al llegar a done emite un evento `final` con el HTML ya
  // renderizado; la página lo aplica, confirma con POST /ack y deja de
  // reconectar porque el servidor va a cerrarse por diseño.
  var RETRY_MIN = 1000;
  var RETRY_MAX = 15000;
  var es = null;
  var retryDelay = RETRY_MIN;
  var reconnectTimer = null;
  var ended = false;

  function scheduleReconnect() {
    if (ended || reconnectTimer) {
      return;
    }
    reconnectTimer = setTimeout(function () {
      reconnectTimer = null;
      connectEvents();
    }, retryDelay);
    retryDelay = Math.min(retryDelay * 2, RETRY_MAX);
  }

  // applyFinal sustituye #content por el fragmento final conservando la lectura.
  function applyFinal(payload) {
    var content = document.getElementById("content");
    if (!content || !payload || typeof payload.html !== "string") {
      return false;
    }
    var state = captureState();
    content.innerHTML = payload.html;
    restoreState(state);
    return true;
  }

  // confirmFinal avisa al servidor de que el estado final ya se aplicó. No
  // modifica ningún documento: el visor sigue siendo de solo lectura.
  function confirmFinal() {
    try {
      fetch("/ack", { method: "POST", headers: { "Accept": "application/json" } })
        .catch(function () { /* el servidor puede cerrar igualmente */ });
    } catch (e) {
      /* el servidor puede cerrar igualmente */
    }
  }

  function finish(payload) {
    ended = true;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (es) {
      es.close();
      es = null;
    }
    setNotice("");
    if (applyFinal(payload)) {
      confirmFinal();
    } else {
      // Último recurso, aún dentro de la gracia: recarga completa.
      window.location.reload();
    }
  }

  function connectEvents() {
    if (ended) {
      return;
    }
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
    // Los eventos con nombre no llegan a onmessage: se registran aparte.
    es.addEventListener("final", function (ev) {
      var payload = null;
      try {
        payload = JSON.parse(ev.data);
      } catch (e) {
        payload = null;
      }
      finish(payload);
    });
    es.onerror = function () {
      if (es) {
        es.close();
        es = null;
      }
      if (ended) {
        return;
      }
      setNotice("Reconectando…");
      scheduleReconnect();
    };
  }

  connectEvents();
})();
